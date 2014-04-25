(function(){

  'use strict';

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
  .directive('binMap', ['bin', function (bin) {
    return {
      // only allow esriMap to be used as an element (<esri-map>)
      restrict: 'E',

      // this directive shares $scope with its parent (this is the default)
      scope: false,

      // define how our template is compiled this gets the $element our directive is on as well as its attributes ($attrs)
      compile: function ($element, $attrs) {
        // remove the id attribute from the main element
        $element.removeAttr('id');

        // append a new div inside this element, this is where we will create our map
        $element.append('<div id=' + $attrs.id + '></div>');

        // since we are using compile we need to return our linker function
        // the 'link' function handles how our directive responds to changes in $scope
        return function (scope, element, attrs, controller) {
          scope.$watch('center', function (newCenter, oldCenter) {
            if (newCenter !== oldCenter) {
              controller.centerAt(newCenter);
            }
          });
        };
      },

      // even though $scope is shared we can declare a controller for manipulating this directive
      // this is great for when you need to expose an API for manipulaiting your directive
      // this is also the best place to setup our map
      controller: function ($scope, $element, $attrs) {
        // setup basemaps
        var streets = L.esri.basemapLayer('Streets');
        var topo = L.esri.basemapLayer('Topographic');
        var oceans = L.esri.basemapLayer('Oceans');
        var natgeo = L.esri.basemapLayer('NationalGeographic');

        var gray = L.layerGroup([
          L.esri.basemapLayer('Gray'),
          L.esri.basemapLayer('GrayLabels')
        ]);
        var darkgray = L.layerGroup([
          L.esri.basemapLayer('DarkGray'),
          L.esri.basemapLayer('DarkGrayLabels')
        ]);
        var imagery = L.layerGroup([
          L.esri.basemapLayer('Imagery'),
          L.esri.basemapLayer('ImageryLabels')
        ]);
        var shadedrelief = L.layerGroup([
          L.esri.basemapLayer('ShadedRelief'),
          L.esri.basemapLayer('ShadedReliefLabels')
        ]);
        var mapattack = L.tileLayer('http://mapattack-tiles-{s}.pdx.esri.com/dark/{z}/{y}/{x}', {
          maxZoom: 18,
          subdomains: '0123'
        });

        var basemaps = {
          'Streets': streets,
          'Topographic': topo,
          'Oceans': oceans,
          'NationalGeographic': natgeo,
          'Gray': gray,
          'DarkGray': darkgray,
          'Imagery': imagery,
          'ShadedRelief': shadedrelief,
          'MapAttack': mapattack
        };

        var basemap = bin.store.get('basemap');

        if (!basemap) {
          basemap = bin.store.set('basemap', 'MapAttack');
        }

        // setup our map options based on the attributes and scope
        var mapOptions = {
          center: ($attrs.center) ? $attrs.center.split(',') : $scope.center,
          zoom: ($attrs.zoom) ? $attrs.zoom : $scope.zoom,
          layers: [basemaps[basemap]]
        };

        var shapeOptions = {
          stroke: true,
          color: '#00b1dc',
          weight: 2,
          opacity: 0.8,
          fill: true,
          fillColor: null,
          fillOpacity: 0.3,
          clickable: true
        };

        // declare our map
        var map = new L.Map($attrs.id, mapOptions);
        var features = L.featureGroup().addTo(map);
        var layers = {};

        L.control.layers(basemaps).addTo(map);

        $scope.toggleGeo = function (id, geo) {
          if (!layers[id]) {
            layers[id] = L.geoJson(geo, {
              style: function (feature) {
                return shapeOptions;
              }
            });
          }
          if (features.hasLayer(layers[id])) {
            return features.removeLayer(layers[id]);
          }
          features.addLayer(layers[id]);
        };

        // lets expose a version of centerAt that takes an array of [lng,lat]
        this.centerAt = function (center) {
          var point = L.latLng(center[0], center[1]);

          map.setView(point);
        };

        // listen for click events and expose them as broadcasts on the scope and suing the scopes click handler
        map.on('click', function (e) {
          // emit a message that bubbles up scopes, listen for it on your scope
          $scope.$emit('map.click', e);

          // use the scopes click fuction to handle the event
          $scope.$apply(function() {
            $scope.click.call($scope, e);
          });
        });

        map.on('baselayerchange', function(e) {
          bin.store.set('basemap', e.name);
        });

        $scope.$on('$destroy', function(){
          map.remove();
        });
      }
    };
  }]);

})();
