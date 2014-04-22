(function(){

  'use strict';

  // Services
  angular.module('Geobin.services', [])

  // App Version
  .value('appVersion', '0.0.0')

  // API Version
  .value('apiVersion', '1')

  // Bins!
  .service('bin', ['$http', '$location', 'apiVersion', function ($http, $location, apiVersion) {

    // Store
    // -----
    // localStorage interface for browser-based user persistence

    var store = this.store = new TinyStore('geobin');

    cleanHistory(store);

    function cleanHistory (store) {
      var h = store.session.history = store.session.history || [];

      for (var i = h.length - 1; i >= 0; i--) {
        var diff = h[i].expires - Math.floor(new Date().getTime() / 1000);
        if (diff < 1) {
          h.splice(i, 1);
        }
      }

      store.save();
    }

    // User
    // ----
    // helpers for browser-based user persistence

    var user = this.user = {};

    user.getHistory = function () {
      return store.get('history');
    };

    // API
    // ---
    // helpers for interacting with the Geobin API

    var api = this.api = {};

    // Create
    // ------
    // POST to /api/{apiVersion}/create
    // expects to get back an object with:
    // * id (string)
    // * expires (unix timestamp)

    api.create = function () {
      $http.post('/api/' + apiVersion + '/create', {})
      .success(function createSuccess (data, status, headers, config) {
        store.session.history.push(data);
        store.save();
        $location.path('/' + data.id);
      });
    };

    // History
    // -------
    // POST to /api/{apiVersion}/history/{binId}
    // expects to get back an object with:
    // * timestamp
    // * headers
    // * body
    // * geojson

    api.history = function (binId, callback) {
      $http.post('/api/' + apiVersion + '/history/' + binId, {})
      .success(function historySuccess (data, status, headers, config) {
        if (status === 200) {
          callback(data);
        }
      });
    };

  }]);

})();
