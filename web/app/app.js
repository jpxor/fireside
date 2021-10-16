class FiresideApp {

    constructor() {
        this.start_keepalive();
    }

    checkStatus(resp) {
        if (resp.status >= 200 && resp.status < 300) return Promise.resolve(resp);
        else return Promise.reject(new Error(resp.status));
    }

    parseJSON(resp) {
        return resp.json();
    }

    start_keepalive() {
        console.log("starting keep-alive loop");
        // send keep-alive signal once every 5 seconds,
        // private server will shutdown after 10 seconds
        // if it does not receive the signal
        var id = -1;
        const send_keepalive = async function() {
            console.log("sending keep-alive");
            fetch("/api/keepalive", {
                method: 'PUT'
            }).then(response => {
                if (response.ok) {
                    return true;
                } else {
                    clearInterval(id);
                    return false;
                }
            }).catch(err => {
                console.error("keep-alive: no response from app server");
                clearInterval(id);
                return false;
            });        
        };
        if (send_keepalive()) {
            id = setInterval(send_keepalive, 5*1000);
        }
    }

    newProfile() {
        let name = prompt('What is your name?');
        if (name == null || name == "") {
            return;
        }
        let encoded_name = encodeURIComponent(name);
        fetch(`/api/v1/users/${encoded_name}`, {
            method: 'POST'
        })
        .then(this.checkStatus)
        .then(_ => {
            this.login(name);
        })
        .catch(err => {
            if (err == "409" || err.message == "409") {
                window.alert("Conflict: name taken");
            } else {
                window.alert("failed to create new user profile: " + err);
            }
        });
    }

    login(username) {
        let encoded_user = encodeURIComponent(username);
        fetch(`/api/v1/auth/${encoded_user}`, {
            method: 'GET'
        })
        .then(this.checkStatus)
        .then(this.parseJSON)
        .then(auth => {
            localStorage.setItem("token", auth.token);
            window.location = `/fireside`;
        })
        .catch(err => {
            if (isNaN(err)) {
                window.location = `/error?code=${encodeURIComponent(err.message)}`;
            } else {
                window.location = `/error?code=${err}`;
            }
        });
    }

    logout() {
        localStorage.removeItem("token");
        window.location = `/welcome`;
    }

}
let app = new FiresideApp();
