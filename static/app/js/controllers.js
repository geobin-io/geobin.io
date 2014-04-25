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

    $scope.isArray = function (obj) {
      return Object.prototype.toString.call(obj) === '[object Array]';
    };

    bin.api.history(binId, function (data) {
      $scope.history = data;
    });
  }]);

})();
