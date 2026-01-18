const errorMessage = document.getElementById("errorMessage");
const successMessage = document.getElementById("successMessage");

if (errorMessage.textContent.trim() !== "") {
  errorMessage.classList.add("show");
}

if (successMessage.textContent.trim() !== "") {
  successMessage.classList.add("show");
}
