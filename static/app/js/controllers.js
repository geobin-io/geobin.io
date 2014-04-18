'use strict';

// Controllers
angular.module('Geobin.controllers', [])

.controller('NavCtrl', ['$scope', '$rootScope', 'BinService', function($scope, $rootScope, BinService) {

  $scope.history = $rootScope.history;

  $scope.newGeobin = BinService.create;

}])

// Home controller
.controller('HomeCtrl', ['$scope', '$rootScope', function($scope, $rootScope) {
  $rootScope.binId = '';
}])

// Bin controller
.controller('BinCtrl', ['$scope', '$rootScope', '$routeParams', function($scope, $rootScope, $routeParams) {
  $scope.binId = $routeParams.binId;
  $rootScope.binId = $routeParams.binId;
}]);
