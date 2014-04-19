(function(){

  'use strict';

  // Services
  angular.module('Geobin.services', [])

  // App Version
  .value('version', '0.0.0')

  // Bins!
  .service('bin', ['$http', '$location', function($http, $location) {

    // Store
    // -----
    // localStorage interface for browser-based user persistence

    var store = this.store = new TinyStore('geobin');
    store.session.history = store.session.history || [];
    store.save();

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

    api.prefix = '/api/1/';

    // Create
    // ------
    // POST to /api/1/history/:binId
    // expects to get back an object with:
    // * id (string)
    // * expires (unix timestamp)

    api.create = function () {
      $http.post(api.prefix + 'create', {})
      .success(function createSuccess (data, status, headers, config) {
        store.session.history.push(data);
        store.save();
        $location.path('/' + data.id);
      });
    };

    // History
    // ------
    // POST to /api/1/history/:binId
    // expects to get back an object with:
    // * id (string)
    // * expires (unix timestamp)

    api.history = function (binId, callback) {
      $http.post(api.prefix + 'history/' + binId, {})
      .success(function historySuccess (data, status, headers, config) {
        if (status === 200) {
          callback(data);
        }
      });
    };

  }]);

})();
