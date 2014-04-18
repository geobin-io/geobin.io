'use strict';

// Services
angular.module('Geobin.services', [])

// App Version
.value('version', '0.0.0')

// Bins!
.service('BinService', ['$rootScope', '$location', function($rootScope, $location){

  var store = new TinyStore('geobin');

  store.session.history = store.session.history || [];
  $rootScope.history = store.session.history;

  this.create = function() {
    var expires = new Date().getTime();
    var id = Math.floor((1 + Math.random()) * expires).toString(16).substring(4);
    var bin = { id: id, expires: expires };

    store.session.history.push(bin);
    store.save();
    $location.path('/' + id);
  };

}]);
