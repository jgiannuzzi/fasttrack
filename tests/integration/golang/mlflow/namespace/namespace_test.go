//go:build integration

package namespace

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api/request"
	"github.com/G-Research/fasttrackml/pkg/api/mlflow/dao/models"
	"github.com/G-Research/fasttrackml/tests/integration/golang/helpers"
)

type NamespaceTestSuite struct {
	helpers.BaseTestSuite
}

func TestNamespaceTestSuite(t *testing.T) {
	suite.Run(t, &NamespaceTestSuite{
		helpers.BaseTestSuite{
			SkipCreateDefaultNamespace: true,
		},
	})
}

func (s *NamespaceTestSuite) Test_Error() {
	tests := []struct {
		name      string
		error     *api.ErrorResponse
		namespace string
	}{
		{
			name:      "RequestNotExistingNamespace",
			error:     api.NewResourceDoesNotExistError("unable to find namespace with code: not-existing-namespace"),
			namespace: "not-existing-namespace",
		},
		{
			name:      "RequestNotExistingDefaultNamespaceExplicitly",
			error:     api.NewResourceDoesNotExistError("unable to find namespace with code: default"),
			namespace: models.DefaultNamespaceCode,
		},
		{
			name:  "RequestNotExistingDefaultNamespaceImplicitly",
			error: api.NewResourceDoesNotExistError("unable to find namespace with code: default"),
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			resp := api.ErrorResponse{}
			s.Require().Nil(
				s.MlflowClient().WithNamespace(
					tt.namespace,
				).WithQuery(
					request.GetExperimentRequest{},
				).WithResponse(
					&resp,
				).DoRequest(
					"%s%s", mlflow.ExperimentsRoutePrefix, mlflow.ExperimentsGetRoute,
				),
			)
			s.Equal(tt.error.Error(), resp.Error())
			s.Equal(api.ErrorCodeResourceDoesNotExist, string(resp.ErrorCode))
		})
	}
}
