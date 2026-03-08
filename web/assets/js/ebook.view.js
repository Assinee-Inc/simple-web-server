const checkboxSelectAll = document.getElementById("checkSelectAll");
const buttonSend = document.getElementById("buttonSend");
const inputsCheckClients = Array.from(document.getElementsByName("clients[]"));
const sendForm = document.getElementById("sendForm");

const myModalAlternative = document.getElementById("staticBackdrop");

checkboxSelectAll.addEventListener("click", toggleSelectClients);

// sendForm.addEventListener("submit", (event) => {
//   event.preventDefault();
//   event.target.submit();
// });

function toggleSelectClients(event) {
  if (event.target.checked) {
    selectAllClients();
  } else {
    unselectAllClients();
  }
  toggleEnableButton(event);
}

function unselectAllClients() {
  inputsCheckClients.forEach((element) => {
    element.checked = false;
  });
}

function selectAllClients() {
  inputsCheckClients.forEach((element) => {
    element.checked = true;
  });
}

function getSelectedInputs() {
  return inputsCheckClients.filter((el) => el.checked);
}

function hasSelectedClients() {
  return getSelectedInputs().length > 0;
}

function openDialogConfirmation() {
  const modalBody = document.querySelector("#staticBackdrop #modal-body-content");
  const list = getSelectedInputs()
    .map((checkbox) => {
      return `<li class="list-group-item"><b>${checkbox.dataset.clientName}</b> - <i>${checkbox.dataset.clientEmail}</i></li>`;
    })
    .join("");
  const totalClientes = getSelectedInputs().length;
  const message = `Você selecionou <b>${totalClientes}</b> ${formatMessage(
    "cliente",
    "clientes",
    totalClientes
  )}. Confirma o envio?`;
  modalBody.innerHTML = `<p>${message}</p><ul class="list-group">${list}</ul>`;
  myModalAlternative.showModal();
}

buttonSend.addEventListener("click", openDialogConfirmation);

function toggleEnableButton(event) {
  if (hasSelectedClients()) {
    if (buttonSend.hasAttribute("disabled")) {
      buttonSend.removeAttribute("disabled");
    }
  } else {
    buttonSend.setAttribute("disabled", "");
    checkboxSelectAll.checked = false;
  }
}

function formatMessage(singular, plural, total) {
  if (total < 0) {
    throw new Error("the total should be greater than 0");
  }
  if (total > 1) {
    return plural;
  }
  return singular;
}
