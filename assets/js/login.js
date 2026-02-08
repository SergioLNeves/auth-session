const form = document.getElementById("loginForm");
const errorMessage = document.getElementById("errorMessage");

form.addEventListener("submit", async (e) => {
  e.preventDefault();

  const submitBtn = form.querySelector('button[type="submit"]');
  submitBtn.disabled = true;
  submitBtn.textContent = "Entrando...";
  hideError();

  try {
    const formData = new FormData(form);

    const response = await fetch("/v1/auth/login", {
      method: "POST",
      body: formData,
    });

    const data = await response.json();

    if (!response.ok) {
      showError(data.detail || "Login failed. Please try again.");
      return;
    }

    window.location.href = "/";
  } catch {
    showError("Connection error. Please try again.");
  } finally {
    submitBtn.disabled = false;
    submitBtn.textContent = "Entrar";
  }
});

function showError(message) {
  errorMessage.textContent = message;
  errorMessage.classList.add("show");
}

function hideError() {
  errorMessage.textContent = "";
  errorMessage.classList.remove("show");
}
