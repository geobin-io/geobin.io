(function(){

  'use strict';

  // Controllers
  angular.module('Geobin.controllers', [])

  .controller('NavCtrl', ['$scope', '$rootScope', 'bin', function($scope, $rootScope, bin) {
    $scope.host = window.location.host;
    $scope.pathname = window.location.pathname.substr(1);
    $scope.bins = bin.store.session.history;

    $scope.$on('$locationChangeSuccess', function(event, args) {
      $scope.pathname = window.location.pathname.substr(1);
    });

    $scope.create = bin.api.create;
  }])

  // Home controller
  .controller('HomeCtrl', ['$scope', 'bin', function($scope, bin) {
    $scope.create = bin.api.create;
    $scope.bins = bin.store.session.history;
  }])

  // Bin controller
  .controller('BinCtrl', ['$scope', '$routeParams', 'bin', function($scope, $routeParams, bin) {
    var binId = $scope.binId = $routeParams.binId;

    $scope.host = window.location.host;

    bin.api.history(binId, function (data) {
      $scope.history = data;
    });
  }]);

})();
