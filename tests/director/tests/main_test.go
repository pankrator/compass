package tests

import (
	"net/http"
	"os"
	"testing"

	"github.com/pkg/errors"

	config "github.com/kyma-incubator/compass/tests/pkg/config"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/machinebox/graphql"
	log "github.com/sirupsen/logrus"

	"github.com/vrischmann/envconfig"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

var (
	conf               = &config.DirectorConfig{}
	dexGraphQLClient   *graphql.Client
	directorHTTPClient *http.Client
)

func TestMain(m *testing.M) {
	dbCfg := persistence.DatabaseConfig{}
	err := envconfig.Init(&dbCfg)
	if err != nil {
		log.Fatal(err)
	}
	tenant.TestTenants.Init()
	defer tenant.TestTenants.Cleanup()

	config.ReadConfig(conf)

	log.Info("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	if err != nil {
		log.Fatal(errors.Wrap(err, "while getting dex token"))
	}
	dexGraphQLClient = gql.NewAuthorizedGraphQLClient(dexToken)
	directorHTTPClient = gql.NewAuthorizedHTTPClient(dexToken)

	exitVal := m.Run()

	os.Exit(exitVal)
}
