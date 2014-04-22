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
  .directive('binMap', [function (version) {
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
            if(newCenter !== oldCenter){
              controller.centerAt(newCenter);
            }
          });
        };
      },

      // even though $scope is shared we can declare a controller for manipulating this directive
      // this is great for when you need to expose an API for manipulaiting your directive
      // this is also the best place to setup our map
      controller: function ($scope, $element, $attrs) {
        // setup our map options based on the attributes and scope
        var mapOptions = {
          center: ($attrs.center) ? $attrs.center.split(',') : $scope.center,
          zoom: ($attrs.zoom) ? $attrs.zoom : $scope.zoom
        };

        // declare our map
        var map = new L.Map($attrs.id, mapOptions);

        L.tileLayer('http://mapattack-tiles-{s}.pdx.esri.com/dark/{z}/{y}/{x}', {
          maxZoom: 18,
          subdomains: '0123'
        }).addTo(map);

        // start exposing an API by setting properties on 'this' which is our controller
        // lets expose the 'addLayer' method so child directives can add themselves to the map
        this.addLayer = function (layer) {
          return map.addLayer(layer);
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

        $scope.$on('$destroy', function(){
          map.remove();
          console.log(map);
        });
      }
    };
  }]);

})();
