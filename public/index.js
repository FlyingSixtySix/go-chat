const ws = new WebSocket('ws://localhost:3000/ws');
ws.addEventListener('error', err => {
    console.error('socket error', err);
});
ws.addEventListener('close', () => {
    console.info('socket closed');
});
ws.addEventListener('message', event => {
    const { type, data } = JSON.parse(event.data);

    switch (type) {
        case 'message':
            // Create the message element
            const li = document.createElement('li');
            const usernameSpan = document.createElement('span');
            const contentSpan = document.createElement('content');

            usernameSpan.className = 'username';
            usernameSpan.textContent = data.username;

            contentSpan.className = 'message';
            contentSpan.textContent = data.content;

            li.append(usernameSpan, contentSpan);

            document.querySelector('.messages').prepend(li);
            break;
        case 'online':
            // Update the online count element
            const onlineCountElement = document.querySelector('.info #online');

            onlineCountElement.textContent = `${data.count} online`;
            break;
    }

    console.log(data);
});

let usernameSet = false;

function sendMessage() {
    const input = document.querySelector('.input-container #input');
    const content = (input.value || '').trim();
    if (content === '') return false;
    if (!usernameSet) {
        if (!setUsername()) return false;
    }
    ws.send(JSON.stringify({
        type: 'message',
        data: {
            content
        }
    }));
    input.value = '';
    return true;
}

function setUsername() {
    const usernameLockButton = document.querySelector('.input-container #username-lock');
    const usernameInput = document.querySelector('.input-container #username');
    const username = (usernameInput.value || '').trim();
    if (username === '') return false;
    ws.send(JSON.stringify({
        type: 'user',
        data: {
            username
        }
    }));
    usernameSet = true;
    usernameInput.disabled = true;
    usernameLockButton.disabled = true;
    return true;
}

document.querySelector('.input-container #input').addEventListener('keydown',
    /**
     * @param {KeyboardEvent} event
     */
    (event) => {
        if (event.key === 'Enter') {
            sendMessage();
        }
    }
);
document.querySelector('.input-container #username').addEventListener('keydown',
    /**
     * @param {KeyboardEvent} event
     */
    (event) => {
        if (event.key === 'Enter') {
            setUsername();
            document.querySelector('.input-container #input').select();
        }
    }
);

document.querySelector('.input-container #send').addEventListener('click', sendMessage);
document.querySelector('.input-container #username-lock').addEventListener('click', setUsername, { once: true });
