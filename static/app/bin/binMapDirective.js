angular.module('Geobin.directives')
.directive('binMap',

  ['store', 'basemaps',

  function (store, basemaps) {

    /**
     * create leaflet layer from geobin request object
     * @param  {Object} obj - geobin request object
     * @param  {Object} body - request body
     * @return {Object} leaflet layer
     */
    function createLayer (obj, body) {
      var content, layer;
      var shapeOptions = {
        color: randomishColor(),
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

        if (obj.path.length) {
          content = valueFromPath(body, obj.path);
        } else {
          content = body;
        }

        layer.bindPopup('<pre>' + JSON.stringify(content, undefined, 2) + '</pre>');
      } else {
        layer = L.geoJson(obj.geo, {
          style: function () {
            return shapeOptions;
          },
          onEachFeature: function (feature, layer) {
            if (body.type === 'FeatureCollection') {
              content = feature;
            } else if (obj.path.length) {
              content = valueFromPath(body, obj.path);
            } else {
              content = body;
            }
            layer.bindPopup('<pre>' + JSON.stringify(content, undefined, 2) + '</pre>');
          }
        });
      }

      return layer;
    }

    function randomishColor () {
      var a = '#'+Math.floor(Math.random()*16777215).toString(16);
      var b = replace(a, 1);
      var c = replace(b, 3);
      var d = replace(c, 5);
      return d;
    }

    function replace (str, index) {
      var c = (Math.floor(Math.random()*10) + 4).toString(16);
      return str.substr(0, index) + c + str.substr(index + 1);
    }

    function valueFromPath (obj, arr) {
      var a = arr.slice(0);
      var k = a.shift();
      if (a.length) {
        return valueFromPath(obj[k], a);
      }
      return obj[k];
    }

    return {
      restrict: 'E',
      scope: false,

      compile: function ($element, $attrs) {
        $element.removeAttr('id');
        $element.append('<div id=' + $attrs.id + '></div>');
        return function (scope, element, attrs, controller) {
          // @TODO: move toggling of geo to a centralized object
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
        var features = {};
        var layers = {};

        $scope.visibleLayers = {};

        L.control.layers(basemaps.all).addTo(map);

        /**
         * toggle a geobin request object's geometry on the map
         * @param  {Object} item - geobin request object
         */
        $scope.toggleGeo = function (item) {
          var color, layer;
          var id = item.timestamp;
          var arr = item.geo;
          var body = JSON.parse(item.body);

          if (!features[id]) {
            features[id] = L.featureGroup();

            for (var i = 0, len = arr.length; i < len; i++) {
              layer = createLayer(arr[i], body);
              features[id].addLayer(layer);
            }
          }

          if (map.hasLayer(features[id])) {
            $scope.visibleLayers[id] = false;
            return map.removeLayer(features[id]);
          }

          $scope.visibleLayers[id] = true;
          map.addLayer(features[id]);
        };

        /**
         * get bounds of all layers in features[] and fit map to total extent
         */
        $scope.zoomToAll = function () {
          var extent;

          for (var obj in features) {
            if (!features.hasOwnProperty(obj)) {
              continue;
            }

            var bounds = features[obj].getBounds();

            if (!extent) {
              extent = L.latLngBounds(bounds);
            } else {
              extent.extend(bounds);
            }
          }

          if (extent) {
            map.fitBounds(extent);
          }
        };

        /**
         * zoom to bounds of geobin request object
         * @param  {Object} item - geobin request object
         */
        $scope.zoomTo = function (item) {
          var id = item.timestamp;
          map.fitBounds(features[id].getBounds());
        };

        map.on('baselayerchange', function(e) {
          store.local.set('basemap', e.name);
        });

        $scope.$on('$destroy', function(){
          map.remove();
        });
      }
    };
  }
]);
