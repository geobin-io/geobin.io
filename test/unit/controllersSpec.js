'use strict';

/* jasmine specs for controllers go here */

describe('controllers', function(){
  var $scope, $location, $rootScope, createController;

  beforeEach(module('Geobin.controllers'));

  beforeEach(inject(function($injector){
    $location = $injector.get('$location');
    $rootScope = $injector.get('$rootScope');
    $scope = $rootScope.$new();

    var $controller = $injector.get('$controller');

    createController = function(name) {
      return $controller(name, {
        '$scope': $scope
      });
    };
  }));

  it('should work?', function() {
    var HomeCtrl = createController('HomeCtrl');
    expect(HomeCtrl).toBeDefined();

    $location.path('/');
    expect($location.path()).toBe('/');
  });
});
