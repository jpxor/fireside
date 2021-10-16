var color = {
    button: "white",
    button_hover: "#fdefdb",
    button_active: "#ffecd1",
    button_text: "#0F57A7",
    disabled: "#94989C",
    disabled_text: "grey"
};
class FiresideButton extends HTMLElement {
    constructor() {
        super();
        let tmpl = document.createElement('template');
        tmpl.innerHTML = `
        <style>
        a.fs-button {
            box-sizing: border-box;
            display: block;
            border-radius: .4em;
            border: none;
            padding: 1em;
            margin: 1em;
            cursor:pointer;
            color: ${color.button_text};
            font-size: large;
            text-decoration: none; 
            background-color: ${color.button};
            transition: all .1s ease-in-out;
            box-shadow: 3px 5px 10px rgba(0, 0, 0, 0.0);
        }
        a.fs-button:hover {
            transform: scale(1.02);
            background-color: ${color.button_hover};
            box-shadow: 3px 5px 10px rgba(0, 0, 0, 0.1);
        }
        a.fs-button:active {
            transform: scale(1);
            background-color: ${color.button_active};
            box-shadow: 3px 5px 10px rgba(0, 0, 0, 0.0);
        }
        a.fs-button.disabled {
            color: ${color.disabled_text};
            cursor:default;
            pointer-events: none;
            background-color: ${color.disabled};
            box-shadow: 3px 5px 10px rgba(0, 0, 0, 0.0);
        }
        </style>
        <div style="text-align: center; box-sizing: border-box;">
            <a id="a" class="fs-button"><slot></slot></a>
        </div>
        `;
        let shadowRoot = this.attachShadow({mode: 'open'});
        shadowRoot.appendChild(tmpl.content.cloneNode(true));
        this.link = this.shadowRoot.getElementById("a");
    }
    static get observedAttributes() {
        return ['href', 'download', "target", "disabled"];
    }
    attributeChangedCallback(name, oldValue, newValue) {
        this.link.setAttribute(name, newValue);
        if (name == "disabled") {
            if (newValue == "") {
                this.link.classList.add("disabled");
            } else {
                this.link.classList.remove("disabled");
            }
        }
    }
}
function register_core_components() {
    window.customElements.define("fs-button", FiresideButton);
}
register_core_components();


class FiresideUserSelection extends HTMLElement {
    constructor() {
        super();
        this.fetchUsers();
        let tmpl = document.createElement('template');
        tmpl.innerHTML = `
        <style>
        * { box-sizing: border-box; }
        .glass-effect {
            background-color: rgba(255, 255, 255, .25);
            -webkit-backdrop-filter: blur(3px);
            backdrop-filter: blur(3px);
            padding: 2em;
        }
        .container {
            border-radius: .4em;
            box-shadow: 2px 2px 10px rgba(0, 0, 0, .1);
            margin: 1em;
            padding: 1em;
            max-height:90vh;
            overflow:auto;
        }
        </style>
        <div class="container glass-effect">
            <slot></slot>
            <div id="user_selector_list"></div>
        </div>
        `;
        let shadowRoot = this.attachShadow({mode: 'open'});
        shadowRoot.appendChild(tmpl.content.cloneNode(true));
        this.viewlist = shadowRoot.getElementById("user_selector_list");
    }
    fetchUsers() {
        this.userPromise = fetch("/api/v1/users");
    }
    render(users, allow_null_users) {
        let checkStatus = resp => {
            if (resp.status >= 200 && resp.status < 300) return Promise.resolve(resp);
            else return Promise.reject(new Error(resp.statusText));
        };
        let parseJSON = resp => {
            return resp.json();
        };
        if ((!users || users == null) && !allow_null_users) {
            if (this.userPromise == null) {
                this.fetchUsers();
            }
            this.userPromise
                .then(checkStatus)
                .then(parseJSON)
                .then(new_users => this.render(new_users, true))
                .catch(err => null);
            return;
        }
        if (users != null) {
            users.sort(function(a,b){
                if (a.name < b.name) return -1;
                if (a.name > b.name) return  1;
                return 0;
            });
            users.forEach(user => {
                console.log(user);
                this.viewlist.innerHTML += `<fs-button href="javascript:app.login('${decodeURIComponent(user.id)}')">${decodeURIComponent(user.name)}</fs-button>`;
            });
        }
        if (this.hasAttribute("allow-new")) {
            this.viewlist.innerHTML += `<fs-button href="javascript:app.newProfile()">New Profile</fs-button>`;
        }
    }
    connectedCallback() {
        this.render();
    }
    attributeChangedCallback() {
        this.render();
    }
}
function register_welcome_page_components() {
    window.customElements.define("fs-user-selection", FiresideUserSelection);
}


class FiresideUserRegister extends HTMLElement {
    constructor() {
        super();
        let tmpl = document.createElement('template');
        tmpl.innerHTML = `
        <style>
        * { box-sizing: border-box; }
        .glass-effect {
            background-color: rgba(255, 255, 255, .25);
            -webkit-backdrop-filter: blur(3px);
            backdrop-filter: blur(3px);
            padding: 2em;
        }
        .container {
            border-radius: .4em;
            box-shadow: 2px 2px 10px rgba(0, 0, 0, .1);
            margin: 1em;
            padding: 1em;
            max-height:90vh;
            overflow:auto;
        }
        label, input {
            display: block;
            width: 94%;
        }
        input {
            border-radius: .4em;
            border:none;
            margin: 0 1em 1em 1em;
            padding: 1em;
        }
        label,p {
            padding: .4em;
            margin: 1em 1em 0 1em;
        }
        </style>
        <div class="container glass-effect">
            <slot></slot>

            <label for="name"><strong>Name</strong></label>
            <input type="text" id="name" name="name"/>

            <label for="password"><strong>Password </strong>(optional*)</label>
            <input type="email" id="password" name="password"/>
            <p>*passwords cannot be recovered if lost</p>

            <fs-button>Start</fs-button>
        </div>
        `;
        let shadowRoot = this.attachShadow({mode: 'open'});
        shadowRoot.appendChild(tmpl.content.cloneNode(true));
    }

}
function register_register_page_components() {
    window.customElements.define("fs-user-register", FiresideUserRegister);
}
