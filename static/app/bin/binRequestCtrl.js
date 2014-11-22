angular.module('Geobin.controllers')
.controller('BinRequestCtrl', ['$scope', '$stateParams', '$location',
  function ($scope, $stateParams, $location) {
    $scope.item = updateItem();
    $scope.requestId = $stateParams.timestamp;
    $scope.$watch('history', function(){
      // @TODO: Clean up. This should ideally be resolved at the router level.
      if ($scope.history === false) {
        $location.path('/' + $stateParams.binId);
        return;
      }
      if ($scope.isEmpty($scope.item)) {
        $scope.item = updateItem();
      }
    });

    function updateItem () {
      if (!$scope.history) {
        return {};
      }
      for (var i = 0, len = $scope.history.length; i < len; i++) {
        var j = parseInt($scope.history[i].timestamp, 10);
        var k = parseInt($stateParams.timestamp, 10);
        if (j && j === k) {
          return $scope.history[i];
        }
      }
    }
  }
]);
