angular.module('Geobin.controllers')
.controller('BinCtrl', ['$scope', '$stateParams', '$location', 'api',
  function ($scope, $stateParams, $location, api) {
    var binId = $scope.binId = $stateParams.binId;
    document.title = 'Geobin | ' + binId;
    $scope.host = window.location.host;
    $scope.startTime = Infinity;

    $scope.isNewReq = function (ts) {
      return ts > $scope.startTime;
    };

    $scope.isArray = function (obj) {
      return Object.prototype.toString.call(obj) === '[object Array]';
    };

    $scope.isEmpty = function (obj) {
      for (var prop in obj) {
        if (obj.hasOwnProperty(prop)) {
          return false;
        }
      }

      return true;
    };

    // boolean flag to see if incoming layer is the very first
    $scope.isFirst = false;

    // bind api.create method to scope for create button
    $scope.create = api.create;

    api.history(binId, function (data) {
      $scope.validBin = (data !== undefined);

      if (!$scope.validBin) {
        $scope.history = false;
        return;
      }

      $scope.startTime = Math.floor(new Date().getTime() / 1000);
      $scope.history = data.reverse();
      if ($scope.history.length === 0) {
        $scope.isFirst = true;
      }
      for (var i = 0; i < data.length; i++) {
        if (data[i].geo) {
          $scope.toggleGeo(data[i]);
        }
      }
      $scope.zoomToAll();
    });

    api.ws.open(binId, function(event) {
      $scope.isNew = true;
      try {
        var data = JSON.parse(event.data);
        $scope.$apply(function(){
          $scope.history.push(data);
          $scope.toggleGeo(data);
          if ($scope.isFirst) {
            $scope.isFirst = false;
            $scope.zoomTo(data);
          }
        });
      } catch (e) {
        console.error('Invalid data received from websocket server');
      }
    });

    $scope.$on('$destroy', function binCtrlDestroy () {
      api.ws.close(binId);
    });
  }
]);
