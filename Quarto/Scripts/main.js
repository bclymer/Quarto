(function (quarto, $) {

    $(function () {
        var canvasElement = $("<canvas />");
        var fullRadius;
        var smallSize = 0.18;
        var farX = 0.8;
        var closeX = 0.266;
        var context = canvasElement.get(0).getContext("2d");
        var selectedPiece = -1;
        $('body').append(canvasElement);
        context.canvas.addEventListener("mousedown", mouseClick, false);
        draw();

        $('body').on("com.bclymer.quarto.userChosePiece", function (event, pieceId) {
            var piece = possiblePieces[pieceId];
            drawPiece(16, piece.square, piece.hole, piece.white, piece.tall);
            selectedPiece = pieceId;
        });

        function mouseClick(event) {
            var x = ((event.x / context.canvas.width) - 0.5) * (context.canvas.width / fullRadius);
            var y = ((event.y / context.canvas.height) - 0.5) * (context.canvas.height / fullRadius);
            if (selectedPiece != -1) {
                for (var i = 0; i < 16; i++) {
                    var coordinates = getLocationXandY(i);
                    if (x > coordinates[0] - smallSize &&
                        x < coordinates[0] + smallSize &&
                        y > coordinates[1] - smallSize &&
                        y < coordinates[1] + smallSize &&
                        Math.sqrt(Math.pow(x - coordinates[0], 2) + Math.pow(y - coordinates[1], 2)) < smallSize) {
                        context.beginPath();
                        context.lineWidth = fullRadius / 95.5;
                        context.strokeStyle = '#003300';
                        var piece = possiblePieces[selectedPiece];
                        boardLocations[i] = selectedPiece;
                        drawPiece(i, piece.square, piece.hole, piece.white, piece.tall);
                        var clearArea = getLocationXandY(16);

                        var adjustedX = context.canvas.width / 2 + (fullRadius * (clearArea[0] - 0.01));
                        var adjustedY = context.canvas.height / 2 + (fullRadius * (clearArea[1] - 0.01));
                        var clearSize = 0.29;
                        var fullSize = clearSize * fullRadius;
                        context.clearRect(adjustedX - fullSize / 2, adjustedY - fullSize / 2, clearSize * fullRadius, clearSize * fullRadius);
                    }
                }
            }
        }

        $(window).resize(function () {
            draw();
        });

        function draw() {
            context.lineWidth = 5;
            context.strokeStyle = '#003300';
            context.canvas.width = window.innerWidth;
            context.canvas.height = window.innerHeight;
            context.clearRect(0, 0, context.canvas.width, context.canvas.height);
            context.fillText(context.canvas.width + "," + context.canvas.height, context.canvas.width - 45, context.canvas.height - 1);
            fullRadius = Math.min(context.canvas.height / 2, context.canvas.width / 2);

            context.beginPath();
            context.lineWidth = fullRadius / 95.5;
            context.strokeStyle = '#003300';
            drawCircle(0, 0, 1);
            for (var i = 0; i < 16; i++) {
                var coordinates = getLocationXandY(i);
                drawCircle(coordinates[0], coordinates[1], smallSize);
            }
            var pieceId = selectedPiece;
            context.stroke();
            for (var i = 0; i < 16; i++) {
                if (boardLocations[i] != -1) {
                    var piece = possiblePieces[boardLocations[i]];
                    drawPiece(i, piece.square, piece.hole, piece.white, piece.tall);
                }
            }
            if (pieceId >= 0) {
                var piece = possiblePieces[pieceId];
                drawPiece(16, piece.square, piece.hole, piece.white, piece.tall);
            }
            selectedPiece = pieceId;
        }

        function drawCircle(x, y, size) {
            var radius = Math.min(context.canvas.height / 2 * size, context.canvas.width / 2 * size);
            var adjustedX = context.canvas.width / 2 + (fullRadius * x);
            var adjustedY = context.canvas.height / 2 + (fullRadius * y);
            context.moveTo(adjustedX + radius, adjustedY);
            context.arc(adjustedX, adjustedY, radius, 0, 2 * Math.PI, false);
        }

        function drawSquare(x, y, size) {
            var adjustedX = context.canvas.width / 2 + (fullRadius * x);
            var adjustedY = context.canvas.height / 2 + (fullRadius * y);
            var fullSize = size * fullRadius;
            context.rect(adjustedX - fullSize / 2, adjustedY - fullSize / 2, fullSize, fullSize);
        }

        function drawPiece(location, square, hole, white, tall) {
            context.beginPath();
            context.lineWidth = fullRadius / 95.5;
            if (tall) {
                context.lineWidth *= 2;
            }
            context.strokeStyle = white ? '#00BB00' : '#002200';
            var coordinates = getLocationXandY(location);
            if (square) {
                drawSquare(coordinates[0], coordinates[1], 0.2);
            } else {
                drawCircle(coordinates[0], coordinates[1], 0.12);
            }
            if (hole) {
                drawCircle(coordinates[0], coordinates[1], 0.08);
            }
            if (!white) {

            }
            context.stroke();
            selectedPiece = -1;
        }

        function getLocationXandY(location) {
            switch (location) {
                case 0:
                    return [0, -farX];
                    break;
                case 1:
                    return [-closeX, -farX + closeX];
                    break;
                case 2:
                    return [closeX, -farX + closeX];
                    break;
                case 3:
                    return [-farX + closeX, -closeX];
                    break;
                case 4:
                    return [0, -closeX];
                    break;
                case 5:
                    return [farX - closeX, -closeX];
                    break;
                case 6:
                    return [-farX, 0];
                    break;
                case 7:
                    return [-closeX, 0];
                    break;
                case 8:
                    return [closeX, 0];
                    break;
                case 9:
                    return [farX, 0];
                    break;
                case 10:
                    return [-farX + closeX, closeX];
                    break;
                case 11:
                    return [0, closeX];
                    break;
                case 12:
                    return [farX - closeX, closeX];
                    break;
                case 13:
                    return [-closeX, farX - closeX];
                    break;
                case 14:
                    return [closeX, farX - closeX];
                    break;
                case 15:
                    return [0, farX];
                    break;
                case 16:
                    return [-0.85, -0.85];
                    break;
            }
        }
    });

    var possiblePieces = [
        {square: true, hole: true, white: true, tall: true},
        {square: true, hole: true, white: true, tall: false},
        {square: true, hole: true, white: false, tall: true},
        {square: true, hole: true, white: false, tall: false},
        {square: true, hole: false, white: true, tall: true},
        {square: true, hole: false, white: true, tall: false},
        {square: true, hole: false, white: false, tall: true},
        {square: true, hole: false, white: false, tall: false},
        {square: false, hole: true, white: true, tall: true},
        {square: false, hole: true, white: true, tall: false},
        {square: false, hole: true, white: false, tall: true},
        {square: false, hole: true, white: false, tall: false},
        {square: false, hole: false, white: true, tall: true},
        {square: false, hole: false, white: true, tall: false},
        {square: false, hole: false, white: false, tall: true},
        { square: false, hole: false, white: false, tall: false }
    ];

    var usedPieces = [];

    var availablePieces = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15];

    var boardLocations = [-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1];

})(window.quarto, jQuery);