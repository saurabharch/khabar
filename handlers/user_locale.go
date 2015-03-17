package handlers

import (
	"net/http"

	"github.com/changer/khabar/db"
	"github.com/changer/khabar/dbapi/user_locale"
	"github.com/changer/khabar/utils"
	"gopkg.in/simversity/gottp.v2"
	gottp_utils "gopkg.in/simversity/gottp.v2/utils"
)

type UserLocale struct {
	gottp.BaseHandler
}

func (self *UserLocale) Put(request *gottp.Request) {
	inputUserLocale := new(user_locale.UserLocale)
	request.ConvertArguments(inputUserLocale)

	if !inputUserLocale.IsValid() {
		request.Raise(gottp.HttpError{http.StatusBadRequest, "user, region_id and language_id must be present."})
		return
	}

	updateParams := make(utils.M)
	updateParams["timezone"] = inputUserLocale.TimeZone
	updateParams["locale"] = inputUserLocale.Locale
	user_locale.Update(db.Conn, inputUserLocale.User, &updateParams)

	request.Raise(gottp.HttpError{http.StatusNoContent, "NoContent"})
}

func (self *UserLocale) Post(request *gottp.Request) {
	userLocale := new(user_locale.UserLocale)
	request.ConvertArguments(userLocale)
	userLocale.PrepareSave()

	if !userLocale.IsValid() {
		request.Raise(gottp.HttpError{http.StatusBadRequest, "user, region_id and language_id must be present."})
		return
	}

	if !utils.ValidateAndRaiseError(request, userLocale) {
		return
	}

	if user_locale.Get(db.Conn, userLocale.User) != nil {
		request.Raise(gottp.HttpError{http.StatusConflict, "User locale information already exists"})
		return
	}

	user_locale.Insert(db.Conn, userLocale)

	request.Raise(gottp.HttpError{http.StatusCreated, string(gottp_utils.Encoder(userLocale))})
}
