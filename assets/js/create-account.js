const form = document.getElementById("createAccountForm");
const errorMessage = document.getElementById("errorMessage");

form.addEventListener("submit", async (e) => {
  e.preventDefault();

  const password = document.getElementById("password").value;
  const confirmPassword = document.getElementById("confirmPassword").value;

  if (password !== confirmPassword) {
    showError("As senhas não coincidem.");
    return;
  }

  const submitBtn = form.querySelector('button[type="submit"]');
  submitBtn.disabled = true;
  submitBtn.textContent = "Criando...";
  hideError();

  try {
    const formData = new FormData(form);

    const response = await fetch("/v1/user/create-account", {
      method: "POST",
      body: formData,
    });

    const data = await response.json();

    if (!response.ok) {
      showError(data.detail || "Erro ao criar conta.");
      return;
    }

    window.location.href = "/";
  } catch {
    showError("Erro de conexão. Tente novamente.");
  } finally {
    submitBtn.disabled = false;
    submitBtn.textContent = "Criar Conta";
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
