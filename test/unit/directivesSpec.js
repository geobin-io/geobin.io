// Mock of leaflet... or at least the pieces of Leaflet needed to run these tests :P
var L = {
  Map: function(id, opt) {
    return {
      on: function(){}
    };
  },
  esri: {
    basemapLayer: function(name) {
      return {};
    }
  },
  layerGroup: function(layers) {
    return {};
  },
  tileLayer: function(url) {
    return {};
  },
  control: {
    layers: function() {
      return {
        addTo: function(map){}
      };
    }
  }
};

describe('directives', function() {
  beforeEach(module('Geobin.services'));
  beforeEach(module('Geobin.directives'));

  describe('bin-map', function() {
    var scope, element;
    var template = '<bin-map id="testMap" center="45.521699,-122.677386" zoom="10"></bin-map>';

    beforeEach(inject(function($rootScope, $compile) {
      scope = $rootScope.$new();
      element = $compile(template)(scope);
      scope.$digest();
    }));

    it('should have a map div in it', function() {
      expect(element.find('div').attr('id')).toBe('testMap');
    });

    // TODO: figure out how to test this s'more.
  });
});
