package cmd

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/G-Research/fasttrackml/pkg/api/admin/service/namespace"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/common"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/config"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/repositories"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/experiment"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/service/run"
	"github.com/G-Research/fasttrackml/pkg/database"
)

var ExampleDataCmd = &cobra.Command{
	Use:   "example-data",
	Short: "Creates example data",
	Long:  "The exampleData command will create example data of a given shape",
	RunE:  exampleDataCmd,
}

func exampleDataCmd(cmd *cobra.Command, args []string) error {
	config := config.NewServiceConfig()

	namespacesAmount := viper.GetInt("namespaces-amount")
	experimentsAmount := viper.GetInt("experiments-amount")
	runsAmount := viper.GetInt("runs-amount")
	paramsAmount := viper.GetInt("params-amount")
	metricsAmount := viper.GetInt("metrics-amount")
	valuesAmount := viper.GetInt("values-amount")

	db, err := database.NewDBProvider(
		config.DatabaseURI,
		config.DatabaseSlowThreshold,
		config.DatabasePoolMax,
	)
	if err != nil {
		return eris.Wrap(err, "error connecting to DB")
	}
	defer db.Close()

	if config.DatabaseReset {
		if err := db.Reset(); err != nil {
			return eris.Wrap(err, "error resetting database")
		}
	}

	if err := database.CheckAndMigrateDB(config.DatabaseMigrate, db.GormDB()); err != nil {
		return eris.Wrap(err, "error running database migration")
	}

	if err := database.CreateDefaultNamespace(db.GormDB()); err != nil {
		return eris.Wrap(err, "error creating default namespace")
	}

	if err := database.CreateDefaultExperiment(db.GormDB(), config.DefaultArtifactRoot); err != nil {
		return eris.Wrap(err, "error creating default experiment")
	}

	if err := database.CreateDefaultMetricContext(db.GormDB()); err != nil {
		return eris.Wrap(err, "error creating default context")
	}

	namespaceService := namespace.NewService(
		config,
		repositories.NewNamespaceRepository(db.GormDB()),
		repositories.NewExperimentRepository(db.GormDB()),
	)
	experimentService := experiment.NewService(
		config,
		repositories.NewTagRepository(db.GormDB()),
		repositories.NewExperimentRepository(db.GormDB()),
	)
	runService := run.NewService(
		repositories.NewTagRepository(db.GormDB()),
		repositories.NewRunRepository(db.GormDB()),
		repositories.NewParamRepository(db.GormDB()),
		repositories.NewMetricRepository(db.GormDB()),
		repositories.NewExperimentRepository(db.GormDB()),
	)
	// artifactService := artifact.NewService(
	// 	repositories.NewRunRepository(db.GormDB()),
	// 	artifactStorageFactory,
	// )

	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	appSignal := make(chan os.Signal, 3)
	signal.Notify(appSignal, os.Interrupt)

	go func() {
		<-appSignal
		stop()
	}()

	for nsIndex := 1; nsIndex <= namespacesAmount; nsIndex++ {
		ns, err := namespaceService.CreateNamespace(
			ctx,
			fmt.Sprintf("example-%d", nsIndex),
			fmt.Sprintf("Example namespace %d", nsIndex),
		)
		if err != nil {
			return eris.Wrapf(err, "error creating example namespace %d", nsIndex)
		}

		for expIndex := 1; expIndex <= experimentsAmount; expIndex++ {
			exp, err := experimentService.CreateExperiment(
				ctx,
				ns,
				&request.CreateExperimentRequest{
					Name: fmt.Sprintf("Example experiment %d", expIndex),
					Tags: []request.ExperimentTagPartialRequest{
						{
							Key:   common.DescriptionTagKey,
							Value: fmt.Sprintf("This is _example_ experiment **%d**", expIndex),
						},
					},
				},
			)
			if err != nil {
				return eris.Wrapf(err, "error creating example experiment %d in namespace %d", expIndex, nsIndex)
			}

			for runIndex := 1; runIndex <= runsAmount; runIndex++ {
				run, err := runService.CreateRun(
					ctx,
					ns,
					&request.CreateRunRequest{
						ExperimentID: fmt.Sprintf("%d", *exp.ID),
						Name:         fmt.Sprintf("Example run %d", runIndex),
						StartTime:    time.Now().UnixMilli(),
						Tags: []request.RunTagPartialRequest{
							{
								Key:   common.DescriptionTagKey,
								Value: fmt.Sprintf("This is _example_ run **%d**", runIndex),
							},
						},
					},
				)
				if err != nil {
					return eris.Wrapf(err, "error creating example run %d in experiment %d in namespace %d",
						runIndex, expIndex, nsIndex)
				}

				params := make([]request.ParamPartialRequest, 0, paramsAmount)
				for paramIndex := 1; paramIndex <= paramsAmount; paramIndex++ {
					params = append(params, request.ParamPartialRequest{
						Key:   fmt.Sprintf("param%d", paramIndex),
						Value: fmt.Sprintf("%f", rand.Float64()),
					})
				}

				ts := time.Now().UnixMilli()
				metrics := make([]request.MetricPartialRequest, 0, metricsAmount*valuesAmount)
				for metricIndex := 1; metricIndex <= metricsAmount; metricIndex++ {
					for valueIndex := 1; valueIndex <= valuesAmount; valueIndex++ {
						metrics = append(metrics, request.MetricPartialRequest{
							Key:       fmt.Sprintf("metric%d", metricIndex),
							Value:     rand.Float64(),
							Timestamp: ts + int64(valueIndex*1000),
							Step:      int64(valueIndex),
						})
					}
				}

				if err := runService.LogBatch(ctx, ns, &request.LogBatchRequest{
					RunID:   run.ID,
					Params:  params,
					Metrics: metrics,
				}); err != nil {
					return eris.Wrapf(err, "error logging params and metrics for example run %d in experiment %d in namespace %d",
						runIndex, expIndex, nsIndex)
				}

				if _, err := runService.UpdateRun(ctx, ns, &request.UpdateRunRequest{
					RunID:   run.ID,
					Status:  string(models.StatusFinished),
					EndTime: time.Now().UnixMilli(),
				}); err != nil {
					return eris.Wrapf(err, "error updating example run %d in experiment %d in namespace %d",
						runIndex, expIndex, nsIndex)
				}

				log.Infof("Created example run %d in experiment %d in namespace %d", runIndex, expIndex, nsIndex)
			}
		}
	}

	return nil
}

// nolint:errcheck,gosec
func init() {
	RootCmd.AddCommand(ExampleDataCmd)

	ExampleDataCmd.Flags().String("default-artifact-root", "./artifacts", "Default artifact root")
	ExampleDataCmd.Flags().String("s3-endpoint-uri", "", "S3 compatible storage base endpoint url")
	ExampleDataCmd.Flags().String("gs-endpoint-uri", "", "Google Storage base endpoint url")
	ExampleDataCmd.Flags().MarkHidden("gs-endpoint-uri")
	ExampleDataCmd.Flags().StringP("database-uri", "d", "sqlite://fasttrackml.db", "Database URI")
	ExampleDataCmd.Flags().Int("database-pool-max", 20, "Maximum number of database connections in the pool")
	ExampleDataCmd.Flags().Duration("database-slow-threshold", 1*time.Second, "Slow SQL warning threshold")
	ExampleDataCmd.Flags().Bool("database-migrate", true, "Run database migrations")
	ExampleDataCmd.Flags().Bool("database-reset", false, "Reinitialize database - WARNING all data will be lost!")
	ExampleDataCmd.Flags().MarkHidden("database-reset")
	ExampleDataCmd.Flags().Int("namespaces-amount", 10, "Number of namespaces to create")
	ExampleDataCmd.Flags().Int("experiments-amount", 10, "Number of experiments to create per namespace")
	ExampleDataCmd.Flags().Int("runs-amount", 10, "Number of runs to create per experiment")
	ExampleDataCmd.Flags().Int("params-amount", 10, "Number of params to create per run")
	ExampleDataCmd.Flags().Int("metrics-amount", 10, "Number of metrics to create per run")
	ExampleDataCmd.Flags().Int("values-amount", 10, "Number of values to create per metric")
}
