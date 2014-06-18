(function(){

  angular.module('Geobin', [
    'ui.router',
    'Geobin.filters',
    'Geobin.services',
    'Geobin.directives',
    'Geobin.controllers',
    'chieffancypants.loadingBar',
    'yaru22.md'
  ])

  .config(['$locationProvider', '$stateProvider', '$urlRouterProvider',
    function ($locationProvider, $stateProvider, $urlRouterProvider) {
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
          templateUrl: '/static/app/partials/home.html'
        })

        // Documentation routes
        .state('api', {
          url: '/doc/api',
          templateUrl: '/static/app/partials/doc/api.html'
        })
        .state('client', {
          url: '/doc/client',
          templateUrl: '/static/app/partials/doc/client.html'
        })
        .state('server', {
          url: '/doc/server',
          templateUrl: '/static/app/partials/doc/server.html'
        })

        // Bin routes
        .state('bin', {
          abstract: true,
          url: '/:binId',
          controller: 'BinCtrl',
          templateUrl: '/static/app/partials/bin.html'
        })
          .state('bin.list', {
            url: '',
            templateUrl: '/static/app/partials/list.html',
            controller: 'BinListCtrl'
          })
          .state('bin.request', {
            url: '/request/:timestamp',
            templateUrl: '/static/app/partials/request.html',
            controller: 'BinRequestCtrl'
          });
    }
  ]);

  // temporary tooltip workaround
  // @TODO: replace jquery/bootstrap elements w/ angularUI
  // http://angular-ui.github.io/bootstrap/
  // $('body').tooltip({
  //     container: 'body',
  //     selector: '[data-toggle=tooltip]'
  // });

})();
