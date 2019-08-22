package persons

import (
  "log"
  "regexp"

  . "github.com/Liquid-Labs/lc-entities-model/go/entities"
  . "github.com/Liquid-Labs/lc-locations-model/go/locations"
  . "github.com/Liquid-Labs/lc-users-model/go/users"
)

var phoneOutFormatter *regexp.Regexp = regexp.MustCompile(`^(\d{3})(\d{3})(\d{4})$`)

const PersonsResourceName = `persons`

// On summary, we don't include address. Note leaving it empty and using
// 'omitempty' on the Person struct won't work because then Persons without
// an address will appear 'incomplete' in the front-end model and never resolve.
type Person struct {
  tableName   struct{} `sql:"persons,select:persons_join_users,alias:person"`
  User
  GivenName   string    `json:"givenName"`
  FamilyName  string    `json:"familyName"`
  Email       string    `json:"email"`
  Phone       string    `json:"phone"`
  BackupEmail string    `json:"backupEmail"`
  BackupPhone string    `json:"backupPhone"`
  AvatarURL   string    `json:"avatarUrl"`
  Addresses   Addresses `json:"addresses" pg:"many2many:address_links,fk:entity_id,joinFK:"`
  ChangeDesc  []string  `json:"changeDesc,omitempty" sql:"-"`
}

func NewPerson(
    name string,
    description string,
    authID string,
    legalID string,
    legalIDType string,
    active bool,
    givenName string,
    familyName string,
    email string,
    phone string,
    backupEmail string,
    backupPhone string,
    avatarURL string,
    addresses Addresses) *Person {
  return &Person{
    struct{}{},
    *NewUser(PersonsResourceName, name, description, authID, legalID, legalIDType, active),
    givenName,
    familyName,
    email,
    phone,
    backupEmail,
    backupPhone,
    avatarURL,
    addresses,
    ([]string)(nil),
  }
}

func (p *Person) FormatOut() *Person {
  p.Phone = phoneOutFormatter.ReplaceAllString(p.Phone, `$1-$2-$3`)
  p.BackupPhone = phoneOutFormatter.ReplaceAllString(p.BackupPhone, `$1-$2-$3`)
  return p
}

func (p *Person) IsConcrete() bool { return true }

func (p *Person) GetEntity() *Entity { return p.User.GetEntity() }

func (p *Person) GetGivenName() string { return p.GivenName }
func (p *Person) SetGivenName(val string) { p.GivenName = val }

func (p *Person) GetFamilyName() string { return p.FamilyName }
func (p *Person) SetFamilyName(val string) { p.FamilyName = val }

func (p *Person) GetEmail() string { return p.Email }
func (p *Person) SetEmail(val string) { p.Email = val }

func (p *Person) GetPhone() string { return p.Phone }
func (p *Person) SetPhone(val string) { p.Phone = val }

func (p *Person) GetBackupEmail() string { return p.BackupEmail }
func (p *Person) SetBackupEmail(val string) { p.BackupEmail = val }

func (p *Person) GetBackupPhone() string { return p.BackupPhone }
func (p *Person) SetBackupPhone(val string) { p.BackupPhone = val }

func (p *Person) GetAvatarURL() string { return p.AvatarURL }
func (p *Person) SetAvatarURL(val string) { p.AvatarURL = val }

func (p *Person) GetAddresses() *Addresses { return &p.Addresses }

func (p *Person) Clone() *Person {
  newChangeDesc := ([]string)(nil)
  if p.ChangeDesc != nil {
    log.Printf("\n\nwhat...\n\n")
    // TODO: should be []*string
    newChangeDesc = make([]string, len(p.ChangeDesc))
    copy(newChangeDesc, p.ChangeDesc)
  }
  log.Printf("\n\nclone: %t\n\n", newChangeDesc == nil)

  return &Person{
    struct{}{},
    *p.User.Clone(),
    p.GivenName,
    p.FamilyName,
    p.Email,
    p.Phone,
    p.BackupEmail,
    p.BackupPhone,
    p.AvatarURL,
    *p.Addresses.Clone(),
    newChangeDesc,
  }
}


func (p *Person) PromoteChanges() {
  log.Printf("\n\npromote: %t", p.ChangeDesc == nil)
  p.ChangeDesc = p.Addresses.PromoteChanges(p.ChangeDesc)
  log.Printf("promote: %t\n\n", p.ChangeDesc == nil)
}
