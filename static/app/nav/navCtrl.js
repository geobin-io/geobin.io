angular.module('Geobin.controllers')
.controller('NavCtrl', ['$scope', '$location', 'api', 'store',
  function ($scope, $location, api, store) {
    $scope.host = window.location.host;
    $scope.pathname = $location.path().substr(1);
    $scope.bins = store.local.session.history;
    $scope.create = api.create;

    $scope.$on('$locationChangeSuccess', function (event, args) {
      $scope.pathname = $location.path().substr(1);
    });
  }
]);
