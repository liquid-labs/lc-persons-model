package persons

import (
  "regexp"

  "github.com/Liquid-Labs/lc-users-model/go/users"
  "github.com/Liquid-Labs/lc-locations-model/go/locations"
)

var phoneOutFormatter *regexp.Regexp = regexp.MustCompile(`^(\d{3})(\d{3})(\d{4})$`)

// On summary, we don't include address. Note leaving it empty and using
// 'omitempty' on the Person struct won't work because then Persons without
// an address will appear 'incomplete' in the front-end model and never resolve.
type PersonSummary struct {
  users.User
  GivenName   string `json:"givenName"`
  FamilyName  string `json:"familyName"`
  Email       string `json:"email"`
  Phone       string `json:"phone"`
  BackupEmail string `json:"backupEmail"`
  BackupPhone string `json:"backupPhone"`
  AvatarURL   string `json:"avatarUrl"`
}

func (p *PersonSummary) FormatOut() {
  p.Phone = phoneOutFormatter.ReplaceAllString(p.Phone, `$1-$2-$3`)
  p.BackupPhone = phoneOutFormatter.ReplaceAllString(p.BackupPhone, `$1-$2-$3`)
}

func (p *PersonSummary) GetGivenName() string { return p.GivenName }
func (p *PersonSummary) SetGivenName(val string) { p.GivenName = val }

func (p *PersonSummary) GetFamilyName() string { return p.FamilyName }
func (p *PersonSummary) SetFamilyName(val string) { p.FamilyName = val }

func (p *PersonSummary) GetEmail() string { return p.Email }
func (p *PersonSummary) SetEmail(val string) { p.Email = val }

func (p *PersonSummary) GetPhone() string { return p.Phone }
func (p *PersonSummary) SetPhone(val string) { p.Phone = val }

func (p *PersonSummary) GetBackupEmail() string { return p.BackupEmail }
func (p *PersonSummary) SetBackupEmail(val string) { p.BackupEmail = val }

func (p *PersonSummary) GetBackupPhone() string { return p.BackupPhone }
func (p *PersonSummary) SetBackupPhone(val string) { p.BackupPhone = val }

func (p *PersonSummary) GetAvatarURL() string { return p.AvatarURL }
func (p *PersonSummary) SetAvatarURL(val string) { p.AvatarURL = val }

func (p *PersonSummary) Clone() *PersonSummary {
  return &PersonSummary{
    *p.User.Clone(),
    p.GivenName,
    p.FamilyName,
    p.Email,
    p.Phone,
    p.BackupEmail,
    p.BackupPhone,
    p.AvatarURL,
  }
}

// We expect an empty address array if no addresses on detail
type Person struct {
  PersonSummary
  Addresses     locations.Addresses  `json:"addresses"`
  ChangeDesc    []string             `json:"changeDesc,omitempty"`
}

func (p *Person) Clone() *Person {
  newChangeDesc := make([]string, len(p.ChangeDesc))
  copy(newChangeDesc, p.ChangeDesc)

  return &Person{
    *p.PersonSummary.Clone(),
    *p.Addresses.Clone(),
    newChangeDesc,
  }
}

func (p *Person) PromoteChanges() {
  p.ChangeDesc = p.Addresses.PromoteChanges(p.ChangeDesc)
}
