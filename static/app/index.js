angular.module('Geobin.controllers', []);
angular.module('Geobin.directives', []);
angular.module('Geobin.filters', []);
angular.module('Geobin.services', []);

angular.module('Geobin', [
  'angulartics',
  'angulartics.google.analytics',
  'chieffancypants.loadingBar',
  'Geobin.controllers',
  'Geobin.directives',
  'Geobin.filters',
  'Geobin.services',
  'yaru22.md',
  'ui.router'
])

.config(['$locationProvider', '$stateProvider', '$urlRouterProvider', '$analyticsProvider',
  function ($locationProvider, $stateProvider, $urlRouterProvider) {

    var basePath = '/static/app/';

    $locationProvider.html5Mode(true);

    // Redirects
    $urlRouterProvider.when(/^\/api(.*)?/, '/doc/api');
    $urlRouterProvider.otherwise('/');

    // Route Configuration
    $stateProvider

      // Index
      .state('home', {
        url: '/',
        controller: 'HomeCtrl',
        templateUrl: basePath + 'home/home.html'
      })

      // Documentation routes
      .state('api', {
        url: '/doc/api',
        templateUrl: basePath + 'doc/api.html'
      })
      .state('client', {
        url: '/doc/client',
        templateUrl: basePath + 'doc/client.html'
      })
      .state('server', {
        url: '/doc/server',
        templateUrl: basePath + 'doc/server.html'
      })

      // Bin routes
      .state('bin', {
        abstract: true,
        url: '/:binId',
        controller: 'BinCtrl',
        templateUrl: basePath + 'bin/bin.html'
      })
        .state('bin.list', {
          url: '',
          templateUrl: basePath + 'bin/binList.html'
        })
        .state('bin.request', {
          url: '/request/:timestamp',
          templateUrl: basePath + 'bin/binRequest.html',
          controller: 'BinRequestCtrl'
        });
  }
])

.run(['$rootScope', 'apiVersion', 'clientVersion',
  function ($rootScope, apiVersion, clientVersion) {
    $rootScope.apiVersion = apiVersion;
    $rootScope.clientVersion = clientVersion;
  }
]);
