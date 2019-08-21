package persons

import (
  "github.com/go-pg/pg/orm"

  "github.com/Liquid-Labs/lc-authentication-api/go/auth"
  . "github.com/Liquid-Labs/lc-entities-model/go/entities"
  . "github.com/Liquid-Labs/lc-locations-model/go/locations"
  . "github.com/Liquid-Labs/terror/go/terror"
  . "github.com/Liquid-Labs/lc-users-model/go/users"
)

func checkAuthentication(db orm.DB) (string, Terror) {
  ctx := db.Context()
  if ctx == nil { return ``, ServerError(`Must set context on db.`, nil) }
  _, authID, err := auth.CheckAuthentication(ctx)

  return authID, err
}

func (p *Person) CreatePersonSelf(db orm.DB) Terror {
  if authID, err := checkAuthentication(db); err != nil {
    return err
  } else {
    q := db.Model(p).Where(`person.auth_id=?`, authID)
    if exists, err := q.Exists(); err != nil {
      return ServerError(`There was a problem verifying person existence.`, err)
    } else if exists {
      return BadRequestError(`Self already exists.`)
    } else {
      im := ConnectItemManager()
      if err := im.CreateRaw(p); err != nil {
        return ServerError(`Problem creating person record.`, err)
      } else {
        for _, address := range *p.GetAddresses() {
          address.EntityID = p.GetID()
        }
        im.CreateRaw(p.GetAddresses())
        return nil
      }
    }
  }
}

func RetrievePersonSelf(db orm.DB) (*Person, Terror) {
  if authID, err := checkAuthentication(db); err != nil {
    return nil, err
  } else {
    p := &Person{}
    q := db.Model(p).Where(`person.auth_id=?`, authID)
    if err := q.Select(); err != nil {
      return nil, ServerError(`There was a problem retrieving person record.`, err)
    } else {
      p.Addresses = make(Addresses, 0)
      p.Addresses.RetrieveByIDRaw(p.GetID(), db)
      // TODO: this is not functionally necessary, but we need exact matches for the tests.
      if p.Addresses != nil && len(p.Addresses) == 0 {
        p.Addresses = Addresses(nil)
      }
      return p.FormatOut(), nil
    }
  }
}

func (p *Person) UpdateSelf(db orm.DB) Terror {
  if authID, err := checkAuthentication(db); err != nil {
    return err
  } else {
    qs := p.UpdateQueries(db)
    for _, q := range qs {
      alias := string(q.GetModel().Table().Alias)
      if alias != `user` {
        q.Join(`users AS "user" ON ?tableAlias.id="user".id`)
      }
      q.Where(`"user".auth_id=?`, authID)
    }

    if err := RunStateQueries(qs, CreateOp); err != nil {
      return ServerError(`Problem updating person record.`, err)
    } else {
      return nil
    }
  }
}

func (p *Person) CreateQueries(db orm.DB) []*orm.Query {
  return append((&p.User).CreateQueries(db),
    db.Model(p).ExcludeColumn(UserFields...) )
}

func (p *Person) deleteMyAddresses(db orm.DB) *orm.Query {
  return db.Model().
    Table(`addresses as "address"`).
    Where(`"address".entity_id=?`, p.GetID())
}

var updateExcludes = make([]string, len(EntityFields))
func init() {
  copy(updateExcludes, UserFields)
  updateExcludes = append(updateExcludes, "id")
}

func (p *Person) UpdateQuerios(db orm.DB) []*orm.Query {
  q := db.Model(p).
    ExcludeColumn(updateExcludes...).
    Where(`"person".id=?id`)
  q.GetModel().Table().SoftDeleteField = nil
  qs := append((&p.User).UpdateQueries(db), q, p.deleteMyAddresses(db))
  return append(qs, p.GetAddresses().CreateQueries(db)...)
}