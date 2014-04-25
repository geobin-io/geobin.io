(function(){

  'use strict';

  // Controllers
  angular.module('Geobin.controllers', [])

  .controller('NavCtrl', ['$scope', '$rootScope', 'bin', function ($scope, $rootScope, bin) {
    $scope.host = window.location.host;
    $scope.pathname = window.location.pathname.substr(1);
    $scope.bins = bin.store.session.history;
    $scope.create = bin.api.create;

    $scope.$on('$locationChangeSuccess', function (event, args) {
      $scope.pathname = window.location.pathname.substr(1);
    });
  }])

  // Home controller
  .controller('HomeCtrl', ['$scope', 'bin', function ($scope, bin) {
    document.title = 'Geobin';
    $scope.create = bin.api.create;
    $scope.bins = bin.store.session.history;
    $scope.enabled = bin.store.enabled;
  }])

  // Bin controller
  .controller('BinCtrl', ['$scope', '$routeParams', 'bin', function ($scope, $routeParams, bin) {
    var binId = $scope.binId = $routeParams.binId;
    document.title = 'Geobin | ' + binId;
    $scope.host = window.location.host;

    // expose a method for handling clicks
    $scope.click = function (e) {
      $scope.center = [e.latlng.lat,e.latlng.lng];
    };

    $scope.isArray = function (obj) {
      return Object.prototype.toString.call(obj) === '[object Array]';
    };

    // listen for click broadcasts
    $scope.$on('map.click', function (event, e) {
      console.log('broadcast', event, e);
    });

    bin.api.history(binId, function (data) {
      $scope.history = data;
    });
  }]);

})();
