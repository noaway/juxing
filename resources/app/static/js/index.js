let index = {
    init: function () {
        document.addEventListener('astilectron-ready', function () {
            index.explore();

            document.getElementById("openlive").addEventListener("click", function () {
                let rid = document.getElementById("rid").value
                astilectron.sendMessage({ "name": rid })
            });

            document.getElementById("clean").addEventListener("click", function () {
                let livepanel = document.getElementById("livepanel");
                livepanel.innerHTML = "";
            });

            let livepanel = document.getElementById("livepanel")
            livepanel.addEventListener('scroll', function (e) {
                if (livepanel.scrollHeight - index.blockheight - livepanel.scrollTop > 50) {
                    index.tobottom = false
                } else {
                    index.tobottom = true
                }
            });
        })
    },
    // .content .block height
    blockheight: 620,
    tobottom: true,
    scroll: function () {
        let livepanel = document.getElementById("livepanel")
        if (index.tobottom) {
            livepanel.scrollTop = livepanel.scrollHeight - index.blockheight
        }
    },
    explore: function () {
        astilectron.onMessage(function (message) {
            let livepanel = document.getElementById("livepanel");
            if (message.type == "神秘人") {
                let span = document.createElement("span");
                span.className = "mdl-chip";
                span.innerHTML = '<span class="mdl-chip__text">' + message.msg + '</span>';
                livepanel.appendChild(span)
                livepanel.appendChild(document.createElement("br"))
            } else if (message.type == "神秘人发礼物") {
                let span = document.createElement("span");
                span.className = "mdl-chip";
                span.innerHTML = '<span class="mdl-chip__text">' + message.msg + '</span>';
                livepanel.appendChild(span)
                livepanel.appendChild(document.createElement("br"))
            } else if (message.type == "error") {
                let div = document.createElement("div");
                div.className = "logo-font";
                div.innerHTML = message.msg;
                livepanel.appendChild(div);
            } else {
                let div = document.createElement("div");
                div.className = "logo-font";
                div.innerHTML = message.msg;
                livepanel.appendChild(div);
            }
            index.scroll()
        });
    }
};
