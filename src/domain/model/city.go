package model

import (
	"github.com/google/uuid"
	"time"
)

type City struct {
	ID        uuid.UUID `json:"id" bun:",pk,nullzero,type:uuid,default:uuid_generate_v4()"`
	Name      string    `bun:"name,unique,notnull"`
	Enabled   bool      `bun:"enabled,notnull,default:true"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:now()"`
}

type СityResp struct {
	Coord Coord  `json:"coord"`
	Name  string `json:"name"`
}

type Coord struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}
