angular.module('Geobin.controllers')
.controller('HomeCtrl', ['$scope', 'api', 'store',
  function ($scope, api, store) {
    document.title = 'Geobin';
    $scope.host = window.location.host;
    $scope.create = api.create;
    $scope.bins = store.local.session.history;
    $scope.enabled = store.local.enabled;
    var binIds = [];
    $scope.counts = {};

    for (var b = 0; b < $scope.bins.length; b++) {
      binIds.push($scope.bins[b].id);
    }

    api.counts(binIds, function(counts) {
      $scope.counts = counts;
    });
  }
]);
