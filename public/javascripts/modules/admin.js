'use strict';

angular.module('AdminModule', [])
.factory('Users', ['$resource', function($resource) {
    return $resource('api/user/:userId', {
        userId: '@_id'
    }, {
      currentUser: {method:'GET'},
      'update': { method:'PUT' }
    });
}])
.controller('AdminCtrl', function ($scope, $rootScope, $http, Users) {

  $rootScope.currentPage = 'admin';
  $scope.users = {}
  $scope.loggedUser = null;

  $scope.find = function() {
    $scope.usersTmp = Users.query();

    $scope.usersTmp.$then(function(data){
      for (var x = 0; x< data.data.length ; x++) {
        $scope.users[data.data[x].id] = data.data[x];
        console.log($scope.users);
      }
    })
    $scope.loggedUser = Users.currentUser({userId: "current"});
  }

  $scope.createUser = function(username, password) {
    if (username && password) {
      
      var user = {};
      user.username = username;
      user.password = password;

      Users.save(user, function(user){
        $scope.message = {clazz: 'alert-success', text: 'User "' + user.username + '" successfully added.'};
        $scope.users[user.id] = user
      },
      function(data) {
        $scope.message = {clazz: 'alert-danger', text: 'Unable to create User ! ' + data};
      });
    }
  }

  $scope.removeUser = function(id) {
    if (id && confirm("Are you sure ?")) {
      Users.remove({userId: id}, function(data){
        $scope.message = {clazz: 'alert-success', text: 'User successfully deleted.'};
        delete($scope.users[id]);
      },
      function(data) {
        $scope.message = {clazz: 'alert-danger', text: 'Unable to delete User ! ' + data.data};
      }
      );
    }
  }

  $scope.editUser = function(index) {
    $scope.editUserUsername = $scope.users[index].username
    $scope.editUserId = $scope.users[index].id
  }

  $scope.updateUser = function (username, password, id) {
    if (username) {
      
      var user = {}
      user.username = username;
      user.id = id;
      user.password = password;

      Users.update({userId: id}, user, function(user){
        $scope.message = {clazz: 'alert-success', text: 'User "' + user.username + '" successfully updated.'};
        $scope.users[user.id] = user;
      },
      function(data){
        $scope.message = {clazz: 'alert-danger', text: 'Unable to update User ! ' + data};
      });
    }
  }

})
;
