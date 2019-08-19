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

func (p *Person) CreateSelf(db orm.DB) Terror {
  if authID, err := checkAuthentication(db); err != nil {
    return err
  } else {
    q := db.Model(p).Where(`person.auth_id=?`, authID)
    if exists, err := q.Exists(); err != nil {
      return ServerError(`Prbloem verify person existence.`, err)
    } else if exists {
      return BadRequestError(`Self already exists.`)
    } else {
      return p.CreateRaw(db)
    }
  }
}

func (p *Person) UpdateSelf(db orm.DB) Terror {
  if authID, err := checkAuthentication(db); err != nil {
    return err
  } else {
    qs := p.UpdateRawQueries(db)
    for _, q := range qs {
      alias := string(q.GetModel().Table().Alias)
      if alias != `user` {
        q.Join(`users AS "user" ON ?tableAlias.id="user".id`)
      }
      q.Where(`"user".auth_id=?`, authID)
    }

    return DoRawUpdate(qs, db)
  }
}

// Create creates (or inserts) a new Person record into the DB without any authorization checks.
func (p *Person) CreateRaw(db orm.DB) Terror {
  if err := (&p.User).CreateRaw(db); err != nil {
    return ServerError(`There was a problem creating the person record.`, err)
  } else {
    q := db.Model(p).ExcludeColumn(UserFields...)
    if _, err := q.Insert(); err != nil {
      return ServerError(`There was a problem creating the user record.`, err)
    } else {
      return p.createAddresses(db)
    }
  }
}

func (p *Person) createAddresses(db orm.DB) Terror {
  if 0 < len(p.GetAddresses()) {
    newAddresses := make([]*EntityAddress, 0)
    for i, address := range p.GetAddresses() {
      addressLink := &EntityAddress{
        EntityID   : p.GetID(),
        LocationID : address.GetLocationID(),
        Idx        : i,
        Label      : address.GetLabel(),
      }
      newAddresses = append(newAddresses, addressLink)
    }
    if _, err := db.Model(&newAddresses).Insert(); err != nil {
      return ServerError(`There was a problem insert address records.`, err)
    }
  }

  return nil
}

var updateExcludes = make([]string, len(EntityFields))
func init() {
  copy(updateExcludes, UserFields)
  updateExcludes = append(updateExcludes, "id")
}

func (p *Person) UpdateRawQueries(db orm.DB) []*orm.Query {
  qs := (&p.User).UpdateRawQueries(db)

  qp := db.Model(p).
    ExcludeColumn(updateExcludes...).
    Where(`"person".id=?id`)
  qp.GetModel().Table().SoftDeleteField = nil

  return append(qs, qp)
}

// Update updates a the person record without any authorization checks.
func (p *Person) UpdateRaw(db orm.DB) Terror {
  if err := DoRawUpdate(p.UpdateRawQueries(db), db); err != nil {
    return err
  } else {
    if _, err := db.Exec(`DELETE FROM entity_addresses WHERE entity_id=?`, p.GetID()); err != nil {
      return ServerError(`Problem updating person addresses.`, err)
    }
    return p.createAddresses(db)
  }
}

// Archive archives a person record withtout any authorization checks.
func (p *Person) ArchiveRaw(db orm.DB) Terror {
  return p.GetEntity().ArchiveRaw(db)
}
