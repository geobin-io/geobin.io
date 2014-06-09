describe('basemaps', function() {
  beforeEach(module('Geobin.services'));

  var basemaps;
  beforeEach(function() {
    inject(function($injector) {
      basemaps = $injector.get('basemaps');
    });
  });

  it('should create some basemaps', function() {
    expect(basemaps.all).toBeDefined();
  });

  it('should have a valid default basemap', function() {
    expect(basemaps.def).toBeDefined();
    expect(basemaps.all[basemaps.def]).toBeDefined();
  });
});

describe('store', function() {
  beforeEach(module('Geobin.services'));

  var store;
  beforeEach(function() {
    inject(function($injector) {
      store = $injector.get('store');
    });
  });

  it('should start with an empty local store', function() {
    expect(store.local).toBeDefined();
    expect(store.local.session.history.length).toBe(0);
  });

  it('should start with a default basemap', function() {
    // TODO: Figure out how to inject the default and test against that.
    expect(store.local.session.basemap).toBe('Streets');
  });

});

describe('api', function() {
  beforeEach(module('Geobin.services'));

  var api, store, $httpBackend;
  beforeEach(inject(function($injector) {
    $httpBackend = $injector.get('$httpBackend');
    $httpBackend.when('POST', '/api/1/create').respond({
      id:'testId',
      expires:new Date().getTime()
    });
    $httpBackend.when('POST', '/api/1/history/testId').respond({
      timestamp: new Date().getTime(),
      headers: [],
      body: 'narf',
      geo: []
    });
    api = $injector.get('api');
    store = $injector.get('store');
  }));

  afterEach(function() {
    $httpBackend.verifyNoOutstandingExpectation();
    $httpBackend.verifyNoOutstandingRequest();
    store.local.clear();
  });

  describe('create', function() {
    it('should post to /api/1/create', function() {
      $httpBackend.expectPOST('/api/1/create');
      api.create();
      $httpBackend.flush();
    });

    it('should create an item in the session history', function() {
      api.create();
      $httpBackend.flush();

      var item = store.local.session.history[0];
      expect(item).toBeDefined();
      expect(item.id).toBe('testId');
    });

    it('should change the $location to the newly created binId', inject(function($location) {
      api.create();
      $httpBackend.flush();

      expect($location.path()).toBe('/testId');
    }));
  });

  describe('history', function() {
    it('should post to /api/1/history/testId', function() {
      $httpBackend.expectPOST('/api/1/history/testId');
      api.history('testId', function(data) { });
      $httpBackend.flush();
    });

    it('should return the history data on success', function() {
      var mock = {
        callback: function(data) {
          expect(data).toBeDefined();
          expect(data.body).toBe('narf');
        }
      };
      spyOn(mock, 'callback').andCallThrough();

      api.history('testId', mock.callback);
      $httpBackend.flush();
      expect(mock.callback).toHaveBeenCalled();
    });

    it('should not run the callback function when given an invalid id', function() {
      $httpBackend.when('POST', '/api/1/history/testMissingId').respond(404);
      $httpBackend.expectPOST('/api/1/history/testMissingId');

      var mock = {
        callback: function(data) {
        }
      };
      spyOn(mock, 'callback');

      api.history('testMissingId', mock.callback);
      $httpBackend.flush();
      expect(mock.callback).not.toHaveBeenCalled();
    });
  });

  // TODO: tests for 'ws' and 'close'
});
