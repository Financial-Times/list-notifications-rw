package main

import (
	"net/http"
	"time"

	"github.com/Financial-Times/list-notifications-rw/bin/list-notifications-rw/resources"
	"github.com/Financial-Times/list-notifications-rw/db"
	"github.com/Financial-Times/list-notifications-rw/mapping"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func server(mapper mapping.NotificationsMapper, nextLink mapping.NextLinkGenerator, db db.DB) {
	r := mux.NewRouter()
	r.HandleFunc("/lists/notifications", resources.ReadNotifications(mapper, nextLink, db))
	r.HandleFunc("/lists/notifications/{uuid}", resources.FilterSyntheticTransactions(resources.WriteNotification(mapper, db))).Methods("PUT")

	r.HandleFunc("/__health", resources.Health(db))
	r.HandleFunc("/__gtg", resources.GTG(db))

	server := &http.Server{
		Handler: r,
		Addr:    ":8080",

		WriteTimeout: 60 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logrus.Info("Starting server on :8080")
	server.ListenAndServe()
}
