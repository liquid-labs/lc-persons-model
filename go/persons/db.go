package persons

import (
  "context"
  "log"
  "github.com/go-pg/pg/orm"

  "github.com/Liquid-Labs/lc-authentication-api/go/auth"
  . "github.com/Liquid-Labs/lc-entities-model/go/entities"
  . "github.com/Liquid-Labs/lc-locations-model/go/locations"
  . "github.com/Liquid-Labs/terror/go/terror"
  . "github.com/Liquid-Labs/lc-users-model/go/users"
)

func requireAuthentication(db orm.DB) (string, Terror) {
  ctx := db.Context()
  if ctx == nil { return ``, ServerError(`Required context not found.`, nil) }

  if authOracle := auth.GetAuthOracleFromContext(ctx); authOracle == nil {
    return ``, ForbiddenError("Request is not authorized.")
  } else {
    return authOracle.GetAuthID(), nil
  }
}

func (p *Person) CreatePersonSelf(ctx context.Context) Terror {
  im := ConnectItemManagerWithContext(ctx)
  if authID, err := requireAuthentication(im.GetDB()); err != nil {
    return err
  } else {
    log.Printf("\n\nabout to start txn")
    if tx, err := im.Begin(); err != nil {
      return ServerError(`There was a problem creating person record.`, err)
    } else {
      q := tx.Model(p).Where(`person.auth_id=?`, authID)
      if exists, err := q.Exists(); err != nil {
        defer im.Rollback() // TODO: check and log error
        return ServerError(`There was a problem verifying person existence.`, err)
      } else if exists {
        defer im.Rollback() // TODO: check and log error
        return BadRequestError(`Self already exists.`)
      } else if err = im.CreateRaw(p); err != nil {
        defer im.Rollback() // TODO: check and log error
        return ServerError(`Problem creating person record.`, err)
      } else {
        for _, address := range *p.GetAddresses() {
          address.OwnerID = p.GetID()
          address.EntityID = p.GetID()
        }
        if err = im.CreateRaw(p.GetAddresses()); err != nil {
          defer im.Rollback()
          return ServerError(`There was a problem creating the person record.`, err)
        } else {
          defer im.Commit()
          return nil
        }
      }
    }
  }
}

func RetrievePersonSelf(db orm.DB) (*Person, Terror) {
  if authID, err := requireAuthentication(db); err != nil {
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
  if authID, err := requireAuthentication(db); err != nil {
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
