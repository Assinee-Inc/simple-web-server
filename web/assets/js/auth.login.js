const togglePassword = document.querySelector("#togglePassword");
const password = document.querySelector("#passwordInput");
const icon = document.querySelector("#toggleIcon");

togglePassword.addEventListener("click", function () {
  const type = password.getAttribute("type") === "password" ? "text" : "password";
  password.setAttribute("type", type);
  icon.classList.toggle("fa-eye");
  icon.classList.toggle("fa-eye-slash");
});
