(function ($) {

	Quarto.register = (function () {
	    function loadRegisterHTML() {
	    	$('body').load('views/register.html', function() {
	    		attachEventsToRegisterPage();
		    });
	    }

		return {
			loadRegisterHTML: loadRegisterHTML
		};
	});

	function attachEventsToRegisterPage() {
		var username = $("#username");

		username.keydown(function (e) {
            if (e.which == 13) {
                $('#submit-username').click();
                return false;
            }
        });
		username.focus();

        $('#register').hide();
        $('#register').slideDown(500);
        $('#submit-username').on("click", function() {
			$.get("validate?uuid=" + username.val(),
				function(data) {
	    			if (data.Valid) {
	    				Quarto.main().loadGameHTML(username.val());
	    			} else {
						toastr.error('Username "' + username.val() + '" is already in use.', 'Error')
	    			}
				},
				"json"
			);
	    });
	}

})(jQuery);