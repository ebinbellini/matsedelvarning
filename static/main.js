"use strict";

// Permissions and registrations
let push_subscribed = false;
let sw_registration = null;


window.onload = main;
function main() {
	register_service_worker();

	const cowe_dance = document.getElementById("dance");
	cowe_dance.classList.add("activated");
	setTimeout(() => {
		move_cow();
		setInterval(move_cow, 1000);
	}, 500);
}

function move_cow() {
	const cowe_dance = document.getElementById("dance");
	const ww = window.innerWidth - cowe_dance.offsetWidth * 1.5;
	const wh = window.innerHeight - cowe_dance.offsetHeight * 1.5;

	cowe_dance.style.left = Math.random() * ww + "px";
	cowe_dance.style.top = Math.random() * wh + "px";
}

function register_service_worker() {
	if ("serviceWorker" in navigator) {
		navigator.serviceWorker.register("/sw.js").then(registration => {
			sw_registration = registration;
			check_if_notifs_are_enabled();
		});
	}
}

function check_if_notifs_are_enabled() {
	sw_registration.pushManager.getSubscription()
		.then(subscription => {
			push_subscribed = !(subscription === null);
		});
}

async function request_notification_permission() {
	const resp = await fetch("/vapid");
	const vapid = await resp.text();
	const applicationServerKey = vapid;
	sw_registration.pushManager.subscribe({
		userVisibleOnly: true,
		applicationServerKey: applicationServerKey
	}).then(subscription => {
		send_subscription_to_server(subscription);
	}).catch(e => {
		console.log(e);
		display_snackbar("Unable to register push notification subscription. Check your connection!")
	});
}

function send_subscription_to_server(subscription) {
	const sub = JSON.stringify(subscription.toJSON());
	fetch("/subscribepush/",
		{
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: sub,
		}).then(resp => {
			if (resp.ok) {
				display_snackbar("Enabled notifications");
			} else {
				resp.text().then(text => {
					display_snackbar(text);
				})
			}
		});
}

// Displays a message at the bottom of the screen for 3 seconds
function display_snackbar(message) {
	let container = document.getElementById("snackbar-container");
	if (!container)
		container = create_snackbar_container();

	const snackbar = create_snackbar(message);
	container.appendChild(snackbar);

	requestAnimationFrame(() =>
		requestAnimationFrame(() => {
			snackbar.classList.add("slideUp");
			setTimeout(() => {
				snackbar.classList.remove("slideUp");
				setTimeout(() => {
					snackbar.parentNode.removeChild(snackbar);
				}, 225);
			}, 3000);
		})
	);
}

function create_snackbar_container() {
	const container = document.createElement("div");
	container.setAttribute("id", "snackbar-container");
	return document.body.appendChild(container);
}

function create_snackbar(message) {
	const snackbar = document.createElement("div");
	snackbar.classList.add("snackbar");
	snackbar.innerHTML = `<span class="snackbar-content"></span>`;
	snackbar.children[0].innerText = message;
	return snackbar;
}