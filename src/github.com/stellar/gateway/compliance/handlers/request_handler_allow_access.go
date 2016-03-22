package handlers

import (
	log "github.com/Sirupsen/logrus"
	"net/http"
	"time"

	"github.com/stellar/gateway/db/entities"
	"github.com/stellar/gateway/protocols/compliance"
	"github.com/stellar/gateway/server"
	"github.com/zenazn/goji/web"
)

func (rh *RequestHandler) HandlerAllowAccess(c web.C, w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	domain := r.PostFormValue("domain")
	publicKey := r.PostFormValue("public_key")
	userId := r.PostFormValue("user_id")

	// TODO check params

	var err error

	if userId != "" {
		entity := &entities.AllowedUser{
			FiName:      name,
			FiDomain:    domain,
			FiPublicKey: publicKey,
			UserId:      userId,
			AllowedAt:   time.Now(),
		}
		err = rh.EntityManager.Persist(entity)
	} else {
		entity := &entities.AllowedFi{
			Name:      name,
			Domain:    domain,
			PublicKey: publicKey,
			AllowedAt: time.Now(),
		}
		err = rh.EntityManager.Persist(entity)
	}

	if err != nil {
		log.WithFields(log.Fields{"err": err}).Warn("Error persisting /allow entity")
		server.Write(w, compliance.InternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
