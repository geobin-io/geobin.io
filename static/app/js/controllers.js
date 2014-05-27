(function(){

  // Controllers
  angular.module('Geobin.controllers', [])

  // Nav controller
  .controller('NavCtrl', ['$scope', '$rootScope', '$location', 'api', 'store', function ($scope, $rootScope, $location, api, store) {
    $scope.host = $location.host();
    $scope.pathname = $location.path().substr(1);
    $scope.bins = store.local.session.history;
    $scope.create = api.create;

    $scope.$on('$locationChangeSuccess', function (event, args) {
      $scope.pathname = $location.path().substr(1);
    });
  }])

  // Home controller
  .controller('HomeCtrl', ['$scope', 'api', 'store', function ($scope, api, store) {
    document.title = 'Geobin';
    $scope.create = api.create;
    $scope.bins = store.local.session.history;
    $scope.enabled = store.local.enabled;
  }])

  // Bin controller
  .controller('BinCtrl', ['$scope', '$routeParams', '$location', 'api', function ($scope, $routeParams, $location, api) {
    var binId = $scope.binId = $routeParams.binId;
    document.title = 'Geobin | ' + binId;
    $scope.host = $location.host();
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

    api.history(binId, function (data) {
      $scope.startTime = Math.floor(new Date().getTime() / 1000);
      $scope.history = data.reverse();
      for (var i = 0; i < data.length; i++) {
        if (data[i].geo) {
          $scope.toggleGeo(data[i]);
        }
      }
    });

    api.ws(binId, function(event) {
      $scope.isNew = true;
      try {
        var data = JSON.parse(event.data);
        $scope.$apply(function(){
          $scope.history.push(data);
          $scope.toggleGeo(data);
        });
      } catch (e) {
        console.error('Invalid data received from websocket server');
      }
    });

    $scope.$on('$destroy', function binCtrlDestroy () {
      api.close(binId);
    });
  }]);

})();
