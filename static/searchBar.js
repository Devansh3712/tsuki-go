function loadUsers(str) {
    var div = document.getElementById("users");
    if (str.length == 0) {
        div.innerHTML = `<p style="color: rgb(130, 130, 130)">No users found.</p>`;
        return;
    }
    $.ajax({
        url: "/search",
        type: "POST",
        data: { search: str },
        success: function(data) {
            if (!data) {
                div.innerHTML = `
                <p style="color: rgb(130, 130, 130)">No users found.</p>`;
                return;
            }
            var content = "";
            data.forEach(function(user) {
                content += `
                <span class="avatar-small">`;
                if (user.Avatar) {
                    content += `<img src="${user.Avatar}" />`;
                } else {
                    content += `<img src="/static/images/avatar.jpg" />`;
                }
                content += `
                </span>
                <a href="/user/${user.Username}">
                    <h3 style="display: inline-block">@${user.Username}</h3>
                </a>
                &nbsp; `;
                if (user.Follows == true) {
                    content += `
                    <button id="follows-${user.Username}" onclick="toggleFollow('${user.Username}')">
                        Unfollow
                    </button>`;
                } else if (user.Follows == false) {
                    content += `
                    <button id="follows-${user.Username}" onclick="toggleFollow('${user.Username}')">
                        Follow
                    </button>`;
                }
                content += `
                    <p class="separator">
                        ${user.Posts} posts &nbsp; ${user.Followers} followers &nbsp; ${user.Following}
                        following
                    </p>`;
            });
            if (data.length == 10) {
                content += `
                <div id="more">
                <h3 style="padding-top: 10px">
                    <a onclick="loadMoreUsers()">
                    <i class="fa-solid fa-circle-chevron-down"></i> More
                    </a>
                </h3>
                </div>`;
            }
            div.innerHTML = content;
        },
    });
}

function toggleFollow(username) {
    var follows = document.getElementById(`follows-${username}`);
    $.ajax({
        url: `/search/${username}/toggle-follow`,
        type: "POST",
        success: function() {
            follows.innerText = follows.innerText == "Unfollow" ? "Follow" : "Unfollow";
        }
    });
}
