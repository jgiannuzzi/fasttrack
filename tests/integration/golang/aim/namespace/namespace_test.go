//go:build integration

package namespace

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/G-Research/fasttrackml/pkg/api/mlflow/api"
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
				s.AIMClient().WithNamespace(
					tt.namespace,
				).WithResponse(
					&resp,
				).DoRequest(
					"/experiments",
				),
			)
			s.Equal(tt.error.Error(), resp.Error())
			s.Equal(api.ErrorCodeResourceDoesNotExist, string(resp.ErrorCode))
		})
	}
}
