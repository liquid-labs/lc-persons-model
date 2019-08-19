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
    locations := make([]*Location, 0)
    for _, address := range p.GetAddresses() {
      location := &Location{
        Address1 : address.GetAddress1(),
        Address2 : address.GetAddress2(),
        City     : address.GetCity(),
        State    : address.GetState(),
        Zip      : address.GetZip(),
      }
      locations = append(locations, location)
    }
    if _, err := db.Model(&locations).Insert(); err != nil {
      return ServerError(`There was a problem inserting the location records.`, err)
    }

    newAddresses := make([]*EntityAddress, 0)
    for i, address := range p.GetAddresses() {
      addressLink := &EntityAddress{
        EntityID   : p.GetID(),
        LocationID : locations[i].GetID(),
        Idx        : i+1, // must start at 1
        Label      : address.GetLabel(),
      }
      newAddresses = append(newAddresses, addressLink)
    }
    if _, err := db.Model(&newAddresses).Insert(); err != nil {
      return ServerError(`There was a problem inserting the address records.`, err)
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
    if _, err := db.Exec(`DELETE FROM l.* FROM locations l LEFT JOIN entity_addresses ea l.location_id=ea.id WHERE ea.id IS NULL`); err != nil {
      // TODO: LOG WARNING; this is not a show stopper, but could cause problems over time
    }
    return p.createAddresses(db)
  }
}

// Archive archives a person record withtout any authorization checks.
func (p *Person) ArchiveRaw(db orm.DB) Terror {
  return p.GetEntity().ArchiveRaw(db)
}
