package persons_test

import (
  // "log"
  "context"
  "os"
  "testing"

  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/require"
  "github.com/stretchr/testify/suite"

  "github.com/Liquid-Labs/lc-authentication-api/go/auth"
  . "github.com/Liquid-Labs/lc-locations-model/go/locations"
  "github.com/Liquid-Labs/lc-rdb-service/go/rdb"
  "github.com/Liquid-Labs/strkit/go/strkit"
  "github.com/Liquid-Labs/terror/go/terror"
  . "github.com/Liquid-Labs/lc-users-model/go/users"
  /* pkg2test */ "github.com/Liquid-Labs/lc-persons-model/go/persons"
)

func init() {
  terror.EchoErrorLog()
}

type PersonIntegrationSuite struct {
  suite.Suite
  Ctx    context.Context
  AuthID string
}
func (s *PersonIntegrationSuite) SetupSuite() {
  s.AuthID = strkit.RandString(strkit.LettersAndNumbers, 12)
  ctx := context.Background()
  authenticator := &auth.Authenticator{}
  authenticator.SetAznID(s.AuthID)
  s.Ctx = context.WithValue(ctx, auth.AuthenticatorKey, authenticator)
}
func TestPersonIntegrationSuite(t *testing.T) {
  if os.Getenv(`SKIP_INTEGRATION`) == `true` {
    t.Skip()
  } else {
    suite.Run(t, new(PersonIntegrationSuite))
  }
}

func (s *PersonIntegrationSuite) TestPersonCreateNoAddresses() {
  p := persons.NewPerson(
    NewUser(`users`, `Bob Woodward`, `a dude`, s.AuthID, `555-55-5555`, `SSN`, true),
    `Robert`,
    `Woodward`,
    `foo@bar.com`,
    `555-555-5555`,
    `baz@bar.com`,
    `555-444-5555`,
    `https://avatars.com/bob`,
    make(Addresses, 0))

  require.NoError(s.T(), p.CreateSelf(rdb.ConnectWithContext(s.Ctx)))
  assert.Equal(s.T(), `Bob Woodward`, p.GetName())
  assert.Equal(s.T(), s.AuthID, p.GetAuthID())
  assert.Equal(s.T(), `Robert`, p.GetGivenName())
  assert.Equal(s.T(), 0, len(p.GetAddresses()))
}
