describe('controllers', function(){
  var $scope, $location, $rootScope, createController;

  beforeEach(module('Geobin.controllers'));
  beforeEach(module('Geobin.services'));

  beforeEach(inject(function($injector){
    $location = $injector.get('$location');
    $rootScope = $injector.get('$rootScope');
    $scope = $rootScope.$new();

    var $controller = $injector.get('$controller');

    createController = function(name, stateParams) {
      return $controller(name, {
        '$scope': $scope,
        '$stateParams': stateParams
      });
    };
  }));

  describe('Home Controller', function() {
    it('should load', function() {
      var HomeCtrl = createController('HomeCtrl');
      expect(HomeCtrl).toBeDefined();
    });

    it('should define some stuff in the $scope', function() {
      var HomeCtrl = createController('HomeCtrl');
      expect($scope.create).toBeDefined();
      expect($scope.bins).toBeDefined();
      expect($scope.enabled).toBeDefined();
      expect($scope.counts).toBeDefined();
    });

  });

  describe('Nav Controller', function() {
    it('should load', function() {
      var NavCtrl = createController('NavCtrl');
      expect(NavCtrl).toBeDefined();
    });

    it('should define some stuff in the $scope', function() {
      var NavCtrl = createController('NavCtrl');
      expect($scope.host).toBeDefined();
      expect($scope.pathname).toBeDefined();
      expect($scope.pathname).toBe('');
      expect($scope.bins).toBeDefined();
      expect($scope.create).toBeDefined();
    });

    it('should update $scope.pathname correctly on location change', function() {
      var NavCtrl = createController('NavCtrl');
      $scope.$on('$locationChangeSuccess', function(event, args) {
        expect($scope.pathname).toBe('test');
      });
      $location.path('/test');
      $scope.$apply();
    });
  });

  describe('Bin Controller', function() {
    // TODO: mock api and inject it everywhere to make these tests moar better.

    it('should load', function() {
      var BinCtrl = createController('BinCtrl', {"binId": 'test'});
      expect(BinCtrl).toBeDefined();
    });

    it('should get the binId from the stateParams', function() {
      var BinCtrl = createController('BinCtrl', {"binId": 'test'});
      expect($scope.binId).toBe('test');
    });

    it('should use the binId in the document.title', function() {
      var BinCtrl = createController('BinCtrl', {"binId": 'test'});
      expect(document.title).toMatch('test');
    });

    it('should fetch the history for the given binId', inject(function(api) {
      // TODO: If we ever upgrade jasmine to 2.0, this needs to change to 'and.callFake'
      spyOn(api, 'history').andCallFake(function(binId, callback) {
        expect(binId).toBe('test');
        // TODO: call the callback with some history data
      });
      var BinCtrl = createController('BinCtrl', {"binId": 'test'});
      expect(api.history).toHaveBeenCalled();
    }));

    // TODO: write tests for the opening and closing of the socket.

    describe('isNewReq', function() {
      it('should calculate whether a request is new or not correctly', function() {
        var BinCtrl = createController('BinCtrl', {"binId": 'test'});
        $scope.startTime = Math.floor(new Date().getTime() / 1000);
        expect($scope.isNewReq($scope.startTime + 10)).toBeTruthy();
        expect($scope.isNewReq($scope.startTime - 10)).toBeFalsy();
      });
    });

    describe('isArray', function() {
      it('should determine if an object is an array or not correctly', function() {
        var BinCtrl = createController('BinCtrl', {"binId": 'test'});
        expect($scope.isArray([1,2,3])).toBeTruthy();
        expect($scope.isArray(undefined)).toBeFalsy();
        expect($scope.isArray({})).toBeFalsy();
        expect($scope.isArray(1)).toBeFalsy();
      });
    });

    describe('isEmpty', function() {
      it('should determine if an object is empty or not correctly', function() {
        var BinCtrl = createController('BinCtrl', {"binId": 'test'});
        expect($scope.isEmpty({})).toBeTruthy();
        expect($scope.isEmpty({"yo": "I ain't empty"})).toBeFalsy();
      });
    });
  });
});
