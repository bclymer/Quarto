(function ($) {
	
	var socket1;
	var socket2;
	var socket3;
	var socket4;
	var socket5;
	var socket6;
	var socket7;

	Quarto.debug = (function () {

		function simulateUsers(userCount) {
			for (var i = 0; i < userCount; i++) {
				new WebSocket('ws://' + window.location.host + '/realtime?username=TestUser' + i);
			}
		}

		return {
			simulateUsers: simulateUsers,
			joinRoom: joinRoom,
			addRoom: addRoom
		}
	});

})(jQuery);