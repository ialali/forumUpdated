document.addEventListener("DOMContentLoaded", function () {
    var loginButton = document.getElementById("loginButton");
    var registerButton = document.getElementById("registerButton");
    var flipper = document.getElementById("flipper");

    loginButton.addEventListener("click", function (event) {
        event.preventDefault(); // Prevent the default link behavior
        flipper.classList.toggle("flip");
    });

    registerButton.addEventListener("click", function (event) {
        event.preventDefault(); // Prevent the default link behavior
        flipper.classList.toggle("flip");
    });
});
