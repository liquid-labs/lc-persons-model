package persons_test

import (
  "database/sql"
  "encoding/json"
  "reflect"
  "strconv"
  "strings"
  "testing"
  "time"

  "github.com/Liquid-Labs/lc-entities-model/go/entities"
  "github.com/Liquid-Labs/lc-locations-model/go/locations"
  "github.com/Liquid-Labs/lc-users-model/go/users"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/require"

  // the package we're testing
  . "github.com/Liquid-Labs/lc-persons-model/go/persons"
)

var trivialPersonSummary = &PersonSummary{
  users.User{
    entities.Entity{
      entities.InternalID(1),
      entities.PublicID(`a`),
      `name`,
      `name is a user`,
      entities.InternalID(1),
      entities.PublicID(`a`),
      false,
      time.Now(),
      time.Now(),
      time.Time{},
    },
    `xzc098`, // AuthID
    `555-55-5555`, // LegalID
    `SSN`, // LegalIDType
    true,
  },
  `GivenName`,
  `FamilyName`,
  `foo@test.com`,
  `555-555-9999`,
  `backup@test.org`,
  `555-555-9998`,
  `http://foo.com/avatar`,
}

func TestPersonSummaryClone(t *testing.T) {
  clone := trivialPersonSummary.Clone()
  assert.Equal(t, trivialPersonSummary, clone, `Original does not match clone.`)
  clone.ID = entities.InternalID(3)
  clone.PubID = entities.PublicID(`b`)
  clone.CreatedAt = clone.CreatedAt.Add(10)
  clone.LastUpdated = clone.LastUpdated.Add(10)
  clone.DeletedAt = time.Now()
  clone.Active = false
  clone.GivenName = `different name`
  clone.FamilyName = `new family`
  clone.Email = `blah@test.com`
  clone.Phone = `555-555-9997`
  clone.BackupEmail = `blah@test.org`
  clone.BackupPhone = `555-555-9996`
  clone.AvatarURL =`http://bar.com/image`

  oReflection := reflect.ValueOf(trivialPersonSummary).Elem()
  cReflection := reflect.ValueOf(clone).Elem()
  for i := 0; i < oReflection.NumField(); i++ {
    assert.NotEqualf(
      t,
      oReflection.Field(i).Interface(),
      cReflection.Field(i).Interface(),
      `Fields '%s' unexpectedly match.`,
      oReflection.Type().Field(i),
    )
  }
}

var trivialPerson = &Person{
  *trivialPersonSummary,
  locations.Addresses{
    &locations.Address{
      locations.Location{
        entities.InternalID(1),
        `a`,
        `b`,
        `c`,
        `d`,
        `e`,
        sql.NullFloat64{2.0, true},
        sql.NullFloat64{3.0, true},
        []string{`f`, `g`},
      },
      1,
      `label a`,
    },
  },
  []string{`h`, `i`},
}

func TestPersonClone(t *testing.T) {
  clone := trivialPerson.Clone()
  assert.Equal(t, trivialPerson, clone, `Original does not match clone.`)
  clone.ID = entities.InternalID(3)
  clone.PubID = entities.PublicID(`b`)
  clone.CreatedAt = clone.CreatedAt.Add(10)
  clone.LastUpdated = clone.LastUpdated.Add(10)
  clone.DeletedAt = time.Now()
  clone.Active = false
  clone.GivenName = `different name`
  clone.FamilyName = `new family`
  clone.Email = `blah@test.com`
  clone.Phone = `555-555-9997`
  clone.BackupEmail = `blah@test.org`
  clone.BackupPhone = `555-555-9996`
  clone.AvatarURL =`http://bar.com/image`
  clone.Addresses = locations.Addresses{
    &locations.Address{
      locations.Location{
        entities.InternalID(2),
        `z`,
        `y`,
        `x`,
        `w`,
        `u`,
        sql.NullFloat64{4.0, true},
        sql.NullFloat64{5.0, true},
        []string{`i`},
      },
      2,
      `label b`,
    },
  }
  clone.ChangeDesc = []string{`j`}

  assert.NotEqual(t, trivialPerson.Addresses, clone.Addresses, `Addresses unexpectedly equal.`)
  aoReflection := reflect.ValueOf(trivialPerson.Addresses[0]).Elem()
  acReflection := reflect.ValueOf(clone.Addresses[0]).Elem()
  for i := 0; i < aoReflection.NumField(); i++ {
    assert.NotEqualf(
      t,
      aoReflection.Field(i).Interface(),
      acReflection.Field(i).Interface(),
      `Fields '%s' unexpectedly match.`,
      aoReflection.Type().Field(i),
    )
  }

  oReflection := reflect.ValueOf(trivialPerson).Elem()
  cReflection := reflect.ValueOf(clone).Elem()
  for i := 0; i < oReflection.NumField(); i++ {
    assert.NotEqualf(
      t,
      oReflection.Field(i).Interface(),
      cReflection.Field(i).Interface(),
      `Fields '%s' unexpectedly match.`,
      oReflection.Type().Field(i),
    )
  }
}

const jdGivenName = "John"
const jdFamilyName = "Doe"
const jdEmail = "johndoe@test.com"
const jdPhone = "555-555-0000"
const jdActive = false

var johnDoeJson string = `
  {
    "givenName": "` + jdGivenName + `",
    "familyName": "` + jdFamilyName + `",
    "email": "` + jdEmail + `",
    "phone": "` + jdPhone + `",
    "active": ` + strconv.FormatBool(jdActive) + `
  }`

var decoder *json.Decoder = json.NewDecoder(strings.NewReader(johnDoeJson))
var johnDoePerson = &Person{}
var decodeErr = decoder.Decode(johnDoePerson)

func TestPersonsDecode(t *testing.T) {
  require.NoError(t, decodeErr, "Unexpected error decoding person JSON.")
  assert.Equal(t, jdGivenName, johnDoePerson.GivenName, "Unexpected display name.")
  assert.Equal(t, jdFamilyName, johnDoePerson.FamilyName, "Unexpected family name.")
  assert.Equal(t, jdEmail, johnDoePerson.Email, "Unexpected email.")
  assert.Equal(t, jdPhone, johnDoePerson.Phone, "Unexpected phone.")
  assert.Equal(t, jdActive, johnDoePerson.Active, "Unexpected active value.")
}

func TestPersonFormatter(t *testing.T) {
  testP := &Person{PersonSummary: PersonSummary{
    Phone: `5555555555`,
    BackupPhone: `1234567890`,
  }}
  testP.FormatOut()
  assert.Equal(t, `555-555-5555`, testP.Phone)
  assert.Equal(t, `123-456-7890`, testP.BackupPhone)
}
