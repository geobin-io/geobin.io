(function(){

  // Directives
  angular.module('Geobin.directives', [])

  // App Version
  .directive('appVersion', ['appVersion', function (version) {
    return function (scope, elm, attrs) {
      elm.text(version);
    };
  }])

  // API Version
  .directive('apiVersion', ['apiVersion', function (version) {
    return function (scope, elm, attrs) {
      elm.text(version);
    };
  }])

  // Bin Map
  .directive('binMap', ['store', 'basemaps', function (store, basemaps) {
    return {
      restrict: 'E',
      scope: false,

      compile: function ($element, $attrs) {
        $element.removeAttr('id');
        $element.append('<div id=' + $attrs.id + '></div>');
        return function (scope, element, attrs, controller) {
          // TODO: move toggling of geo to a centralized object
          // this will allow all shapes to be on by default
          // and incoming geojson from websockets to be added on the fly
          // scope.$watch('geo', function (newCenter, oldCenter) {});
        };
      },

      controller: function ($scope, $element, $attrs) {
        var current = store.local.get('basemap');

        // this is a hack to invalidate the cached basemaps
        // leaflet seems to be unable to initialize a map with a previously existing basemap
        // after another map has been destroyed
        basemaps.init();

        if (!current) {
          current = store.local.set('basemap', basemaps.def);
        }

        var mapOptions = {
          center: ($attrs.center) ? $attrs.center.split(',') : $scope.center,
          zoom: ($attrs.zoom) ? $attrs.zoom : $scope.zoom,
          layers: [basemaps.all[current]]
        };

        // declare our map
        var map = new L.Map($attrs.id, mapOptions);
        var features = [];
        var layers = {};

        L.control.layers(basemaps.all).addTo(map);

        function randomColor () {
          return '#'+Math.floor(Math.random()*16777215).toString(16);
        }

        $scope.toggleGeo = function (id, arr) {
          var color, layer;

          if (!features[id]) {
            features[id] = L.featureGroup();

            for (var i = 0; i < arr.length; i++) {
              layer = createLayer(arr[i]);
              console.log(layer);

              features[id].addLayer(layer);
            }
          }

          if (map.hasLayer(features[id])) {
            return map.removeLayer(features[id]);
          }

          map.addLayer(features[id]);
        };

        function createLayer (obj) {
          var layer;
          var popup = '<pre>' + JSON.stringify(obj.geo, undefined, 2) + '</pre>';
          var shapeOptions = {
            color: randomColor(),
            stroke: true,
            weight: 2,
            opacity: 0.8,
            fill: true,
            fillColor: null,
            fillOpacity: 0.3,
            clickable: true
          };

          if (obj.radius) {
            layer = L.circle(
              obj.geo.coordinates.reverse(),
              obj.radius,
              shapeOptions
            );
          } else {
            layer = L.geoJson(obj.geo, {
              style: function () {
                return shapeOptions;
              }
            });
          }

          layer.bindPopup(popup);
          return layer;
        }

        map.on('baselayerchange', function(e) {
          store.local.set('basemap', e.name);
        });

        $scope.$on('$destroy', function(){
          map.remove();
        });
      }
    };
  }]);

})();
