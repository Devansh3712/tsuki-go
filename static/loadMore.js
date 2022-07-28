// Load more feed posts
function loadMoreFeed() {
    $.ajax({
        url: "/feed/more",
        type: "GET",
        success: function (data) {
            data.forEach(function (post) {
                content = `<span class="avatar-small">`;
                if (post.Avatar) {
                    content += `<img src="${post.Avatar}" />`;
                } else {
                    content += `<img src="/static/images/avatar.jpg" />`;
                }
                content += `</span>
                <h3 style="display: inline-block">
                    <a href="/user/${post.Username}">@${post.Username}</a>
                </h3>
                <a href="/post/${post.Id}">
                    <p>${post.Body}</p>
                    <p class="separator">${post.CreatedAt}</p>
                </a>`;
                $("#posts").append(content);
            });
        },
    });
}

// Load more comments on a post
function loadMoreComments(postId) {
    $.ajax({
        url: `/post/${postId}/comments`,
        type: "GET",
        success: function (data) {
            data.forEach(function (comment) {
                content = `<p>${comment.Body}</p>
                    <p class="separator">
                    <a href="/user/${comment.Username}">@${comment.Username}</a> &nbsp;`;
                if (comment.Self) {
                    content += `<a href="/post/${postId}/comment/delete?commentId=${comment.Id}">
                        <i class="fa-regular fa-trash-can"></i> Delete
                    </a>`;
                }
                content += `</p>`;
                $("#comments").append(content);
            });
        },
    });
}

// Load more users in search
function loadMoreUsers() {
    $.ajax({
        url: "/search/more",
        type: "GET",
        success: function (data) {
            data.forEach(function (user) {
                content = `<form
                name="follow"
                action="/search/{{ .Username }}/toggle-follow"
                method="POST"
                enctype="multipart/form-data"
                >
                <span class="avatar-small">`;
                if (user.Avatar) {
                    content += `<img src="${user.Avatar}" />`;
                } else {
                    content += `<img src="/static/images/avatar.jpg" />`;
                }
                content += `</span>
                <a href="/user/${user.Username}">
                    <h3 style="display: inline-block">@${user.Username}</h3>
                </a>
                &nbsp; `;
                if (user.Follows == true) {
                    content += `<button type="submit">Unfollow</button>`;
                } else if (user.Follows == false) {
                    content += `<button type="submit">Follow</button>`;
                }
                content += `</form>
                <p class="separator">
                    ${user.Posts} posts &nbsp; ${user.Followers} followers &nbsp; ${user.Following}
                    following
                </p>`;
                $("#users").append(content);
            });
        },
    });
}

// Load more posts of a user
function loadMorePosts(username) {
    $.ajax({
        url: `/user/${username}/posts/more`,
        type: "GET",
        success: function (data) {
            data.forEach(function (post) {
                $("#posts").append(`<a href="/post/${post.Id}">
                    <p class="content">${post.Body}</p>
                    <p class="separator">${post.CreatedAt}</p>
                </a>`);
            });
        },
    });
}
