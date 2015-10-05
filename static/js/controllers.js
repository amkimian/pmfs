
angular.module('pmfs.controllers', []);

angular.module('pmfs.controllers').service('fsService', function($http) {
    this.getData = function(path) {
      return $http({ method: 'GET', url: 'http://localhost:5967' + path});
    };
});

angular.module('pmfs.controllers').controller('pmfsController', function($scope, fsService) {
   $scope.test = "This is a test";
   $scope.path = "/";
   $scope.getData = function() {
     fsService.getData($scope.path).then(function(data, status, config, headers) {
       $scope.data = data.data;
     }
   )};
   $scope.jumpInto = function(name) {
     $scope.path = $scope.path + name;
     $scope.getData();
   };
});
