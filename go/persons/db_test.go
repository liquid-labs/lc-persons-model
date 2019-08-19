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
func (s *PersonIntegrationSuite) SetupTest() {
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

func (s *PersonIntegrationSuite) TestPersonCreateWithAddresses() {
  as := make(Addresses, 0)
  as = append(as,
    &Address{
      Location: *NewLocation(`100 Main Str`, ``, `Pflugerville`, `TX`, `78745`),
      Idx: 1,
      Label: `Home`,
    },
    &Address{
      Location: *NewLocation(`221 Baker Str`, `#B`, `London`, `AZ`, `654321`),
      Idx: 2,
      Label: `Vacation`,
    })
  p := persons.NewPerson(
    NewUser(`users`, `Address Woman`, `a lady`, s.AuthID, `555-44-5555`, `SSN`, true),
    `Address`,
    `Woman`,
    `blah@bar.com`,
    `555-333-5555`,
    `flop@bar.com`,
    `555-222-5555`,
    `https://avatars.com/address`,
    as)

  require.NoError(s.T(), p.CreateSelf(rdb.ConnectWithContext(s.Ctx)))
  assert.Equal(s.T(), `Address Woman`, p.GetName())
  assert.Equal(s.T(), s.AuthID, p.GetAuthID())
  assert.Equal(s.T(), `Address`, p.GetGivenName())
  assert.Equal(s.T(), 2, len(p.GetAddresses()))
  a1 := p.GetAddresses()[0]
  assert.Equal(s.T(), `100 Main Str`, a1.GetAddress1())
  assert.Equal(s.T(), ``, a1.GetAddress2())
  assert.Equal(s.T(), `Pflugerville`, a1.GetCity())
  assert.Equal(s.T(), `TX`, a1.GetState())
  assert.Equal(s.T(), `78745`, a1.GetZip())
  a2 := p.GetAddresses()[1]
  assert.Equal(s.T(), `#B`, a2.GetAddress2())
}
