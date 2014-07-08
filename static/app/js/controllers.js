(function(){

  // Controllers
  angular.module('Geobin.controllers', [])

  // Nav controller
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
  ])

  // Home controller
  .controller('HomeCtrl', ['$scope', 'api', 'store', function ($scope, api, store) {
    document.title = 'Geobin';
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
  }])

  // Bin controller
  .controller('BinCtrl', ['$scope', '$stateParams', '$location', 'api', function ($scope, $stateParams, $location, api) {
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
      if ($scope.validBin) {
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
      }
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
  }])

  .controller('BinListCtrl', ['$scope', function ($scope) {}])

  .controller('BinRequestCtrl', ['$scope', '$stateParams',
    function ($scope, $stateParams) {
      $scope.item = updateItem();
      $scope.$watch('history', function(){
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

})();
