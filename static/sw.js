self.addEventListener('push', event => {
	event.waitUntil(handlePush(event));
});

async function handlePush(event) {
	const text = event.data.text();

	const options = {
		body: text,
		actions: [],
	};

	return self.registration.showNotification(text, options);
}

self.addEventListener('pushsubscriptionchange', event => {
	event.waitUntil(swRegistration.pushManager.subscribe(event.oldSubscription.options)
		.then(subscription => {
			return fetch("/updatepush", {
				method: "POST",
				headers: {
					"Content-type": "application/json"
				},
				body: JSON.stringify({
					old_endpoint: event.oldSubscription ? event.oldSubscription.endpoint : null,
					new_subscription: subscription.toJSON(),
				})
			})
		})
	);
});
