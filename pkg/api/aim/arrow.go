package aim

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/apache/arrow/go/v15/arrow/ipc"
	"github.com/apache/arrow/go/v15/arrow/memory"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/G-Research/fasttrackml/pkg/api/aim/query"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/controller"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/common/middleware/namespace"
	"github.com/G-Research/fasttrackml/pkg/database"
)

// SearchRunsArrow is a handler for the /search/runs/arrow endpoint.
// TODO:get back and fix `gocyclo` problem.
//
//nolint:gocyclo
func SearchRunsArrow(c *fiber.Ctx) error {
	ns, err := namespace.GetNamespaceFromContext(c.Context())
	if err != nil {
		return api.NewInternalError("error getting namespace from context")
	}

	q := struct {
		Query  string `query:"q"`
		Limit  int    `query:"limit"`
		Offset string `query:"offset"`
		// TODO skip_system is unused - should we keep it?
		SkipSystem     bool `query:"skip_system"`
		ReportProgress bool `query:"report_progress"`
		ExcludeParams  bool `query:"exclude_params"`
		ExcludeTraces  bool `query:"exclude_traces"`
	}{}

	if err := c.QueryParser(&q); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if c.Query("report_progress") == "" {
		q.ReportProgress = true
	}

	tzOffset, err := strconv.Atoi(c.Get("x-timezone-offset", "0"))
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "x-timezone-offset header is not a valid integer")
	}

	qp := query.QueryParser{
		Default: query.DefaultExpression{
			Contains:   "run.archived",
			Expression: "not run.archived",
		},
		Tables: map[string]string{
			"runs":        "runs",
			"experiments": "Experiment",
		},
		TzOffset:  tzOffset,
		Dialector: database.DB.Dialector.Name(),
	}
	pq, err := qp.Parse(q.Query)
	if err != nil {
		return err
	}

	var total int64
	if tx := database.DB.
		Model(&database.Run{}).
		Count(&total); tx.Error != nil {
		return fmt.Errorf("unable to count total runs: %w", tx.Error)
	}

	log.Debugf("Total runs: %d", total)

	tx := database.DB.
		InnerJoins(
			"Experiment",
			database.DB.Select(
				"ID", "Name",
			).Where(
				&models.Experiment{NamespaceID: ns.ID},
			),
		).
		Order("row_num DESC")

	if q.Limit > 0 {
		tx.Limit(q.Limit)
	}

	if q.Offset != "" {
		run := &database.Run{
			ID: q.Offset,
		}
		// TODO:DSuhinin -> do we need `namespace` restriction? it seems like yyyyess, but ....
		if err := database.DB.
			Select("row_num").
			First(&run).
			Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("unable to find search runs offset %q: %w", q.Offset, err)
		}

		tx.Where("row_num < ?", run.RowNum)
	}

	if !q.ExcludeParams {
		tx.Preload("Params")
		tx.Preload("Tags")
	}

	if !q.ExcludeTraces {
		tx.Preload("LatestMetrics")
	}

	var runs []database.Run
	pq.Filter(tx).Find(&runs)
	if tx.Error != nil {
		return fmt.Errorf("error searching runs: %w", tx.Error)
	}

	log.Debugf("Found %d runs", len(runs))

	run_ids := make([]string, len(runs))
	for i, r := range runs {
		run_ids[i] = r.ID
	}

	var paramKeys []string
	var tagKeys []string
	var metricKeys []string

	if !q.ExcludeParams {
		if err := database.DB.
			Model(&database.Param{}).
			Where("run_uuid IN ?", run_ids).
			Distinct("key").
			Order("key").
			Find(&paramKeys).
			Error; err != nil {
			return fmt.Errorf("error finding param keys: %w", err)
		}

		if err := database.DB.
			Model(&database.Tag{}).
			Where("run_uuid IN ?", run_ids).
			Distinct("key").
			Order("key").
			Find(&tagKeys).
			Error; err != nil {
			return fmt.Errorf("error finding tag keys: %w", err)
		}
	}

	if !q.ExcludeTraces {
		if err := database.DB.
			Model(&database.LatestMetric{}).
			Where("run_uuid IN ?", run_ids).
			Distinct("key").
			Order("key").
			Find(&metricKeys).
			Error; err != nil {
			return fmt.Errorf("error finding metric keys: %w", err)
		}
	}

	c.Set("Content-Type", "application/vnd.apache.arrow.stream")
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		start := time.Now()
		if err := func() error {
			pool := memory.NewGoAllocator()

			fields := []arrow.Field{
				{Name: "run_id", Type: arrow.BinaryTypes.String},
				{Name: "name", Type: arrow.BinaryTypes.String},
				{Name: "experiment_name", Type: arrow.BinaryTypes.String},
				{Name: "creation_time", Type: arrow.FixedWidthTypes.Timestamp_ms},
				{Name: "end_time", Type: arrow.FixedWidthTypes.Timestamp_ms},
				{Name: "archived", Type: arrow.FixedWidthTypes.Boolean},
				{Name: "active", Type: arrow.FixedWidthTypes.Boolean},
			}
			metricIndex := len(fields)
			if !q.ExcludeTraces {
				for _, k := range metricKeys {
					fields = append(fields, arrow.Field{Name: "metrics:" + k, Type: arrow.PrimitiveTypes.Float64})
				}
			}
			paramIndex := len(fields)
			if !q.ExcludeParams {
				for _, k := range paramKeys {
					fields = append(fields, arrow.Field{Name: "params:" + k, Type: arrow.BinaryTypes.String})
				}
			}
			tagIndex := len(fields)
			if !q.ExcludeParams {
				for _, k := range tagKeys {
					fields = append(fields, arrow.Field{Name: "tags:" + k, Type: arrow.BinaryTypes.String})
				}
			}
			schema := arrow.NewSchema(fields, nil)

			writer := ipc.NewWriter(w, ipc.WithAllocator(pool), ipc.WithSchema(schema))
			//nolint:errcheck
			defer writer.Close()

			b := array.NewRecordBuilder(pool, schema)
			defer b.Release()

			for i, r := range runs {
				b.Field(0).(*array.StringBuilder).Append(r.ID)
				b.Field(1).(*array.StringBuilder).Append(r.Name)
				b.Field(2).(*array.StringBuilder).Append(r.Experiment.Name)
				b.Field(3).(*array.TimestampBuilder).Append(arrow.Timestamp(r.StartTime.Int64))
				b.Field(4).(*array.TimestampBuilder).Append(arrow.Timestamp(r.EndTime.Int64))
				b.Field(5).(*array.BooleanBuilder).Append(r.LifecycleStage == database.LifecycleStageDeleted)
				b.Field(6).(*array.BooleanBuilder).Append(r.Status == database.StatusRunning)

				if !q.ExcludeTraces {
					metrics := make(map[string]float64, len(r.LatestMetrics))
					for _, m := range r.LatestMetrics {
						v := m.Value
						if m.IsNan {
							v = math.NaN()
						}
						metrics[m.Key] = v
					}
					for i, k := range metricKeys {
						if v, ok := metrics[k]; ok {
							b.Field(i + metricIndex).(*array.Float64Builder).Append(v)
						} else {
							b.Field(i + metricIndex).(*array.Float64Builder).AppendNull()
						}
					}
				}

				if !q.ExcludeParams {
					params := make(map[string]string, len(r.Params))
					for _, p := range r.Params {
						params[p.Key] = p.Value
					}
					for i, k := range paramKeys {
						if v, ok := params[k]; ok {
							b.Field(i + paramIndex).(*array.StringBuilder).Append(v)
						} else {
							b.Field(i + paramIndex).(*array.StringBuilder).AppendNull()
						}
					}

					tags := make(map[string]string, len(r.Tags))
					for _, t := range r.Tags {
						tags[t.Key] = t.Value
					}
					for i, k := range tagKeys {
						if v, ok := tags[k]; ok {
							b.Field(i + tagIndex).(*array.StringBuilder).Append(v)
						} else {
							b.Field(i + tagIndex).(*array.StringBuilder).AppendNull()
						}
					}
				}

				if (i+1)%1000 == 0 {
					if err := controller.WriteStreamingRecord(writer, b.NewRecord()); err != nil {
						return fmt.Errorf("error writing record: %w", err)
					}
				}
			}

			if b.Field(0).Len() > 0 {
				if err := controller.WriteStreamingRecord(writer, b.NewRecord()); err != nil {
					return fmt.Errorf("error writing record: %w", err)
				}
			}

			return nil
		}(); err != nil {
			log.Errorf("Error encountered in %s %s: error streaming runs: %s", c.Method(), c.Path(), err)
		}

		log.Infof("body - %s %s %s", time.Since(start), c.Method(), c.Path())
	})

	return nil
}
