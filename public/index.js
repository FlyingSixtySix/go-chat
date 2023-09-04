const ws = new WebSocket('ws://localhost:3000/ws');
ws.onerror = err => {
    console.error(err);
};
ws.onmessage = message => {
    const data = JSON.parse(message.data);

    const li = document.createElement('li');
    const usernameSpan = document.createElement('span');
    const contentSpan = document.createElement('span');

    usernameSpan.className = 'username';
    usernameSpan.textContent = data.username;

    contentSpan.className = 'message';
    contentSpan.textContent = data.message;

    li.append(usernameSpan, contentSpan);

    document.querySelector('.messages').appendChild(li);
};

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

document.querySelector('.input-container .send').addEventListener('click', sendMessage);

function sendMessage() {
    const username = (document.querySelector('#username').value || '').trim();
    const input = document.querySelector('#input');
    const message = (input.value || '').trim();
    if (message === '') return;
    ws.send(JSON.stringify({
        username,
        message
    }));
    input.value = '';
}