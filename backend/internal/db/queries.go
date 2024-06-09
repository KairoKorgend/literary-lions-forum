package dbserver

// PostQuerie TODO:
// COALESCE(u.avatar_path, 'default_avatar.png') AS avatar_path
// COALESCE(a.id, 0) AS attachment_id
// COALESCE(a.attachment_path, '') AS attachment_path
// LEFT JOIN
//     attachments a ON p.id = a.post_id

var CategoryQuery = `
SELECT 
    c.title, 
    s.id, 
    s.title, 
    s.icon_path 
FROM 
    categories c 
JOIN 
    subcategories s ON c.id = s.category_id
`

var CategoryPostQuery = `
SELECT 
p.id,
p.subcategory_id, 
p.author AS username,
u.id,
u.profile_picture_path AS user_image,
p.event_time, 
p.title,
p.content, 
COALESCE(l.like_count, 0) AS likes, 
COALESCE(d.dislike_count, 0) AS dislikes
FROM 
    posts p 
JOIN 
    users u ON p.author_id = u.id 
LEFT JOIN 
    (SELECT post_id, COUNT(*) AS like_count FROM likes GROUP BY post_id) l ON p.id = l.post_id 
LEFT JOIN 
    (SELECT post_id, COUNT(*) AS dislike_count FROM dislikes GROUP BY post_id) d ON p.id = d.post_id 
LEFT JOIN 
    (SELECT post_id, COUNT(*) AS comment_count FROM comments GROUP BY post_id) c ON p.id = c.post_id
WHERE 
    p.subcategory_id = ?
ORDER BY 
    %s %s;
`

var IndexPostQuery = `
SELECT 
    p.id,
    p.subcategory_id, 
    p.author AS username,
    u.id,
    u.profile_picture_path AS user_image,
    p.event_time,
    p.title,
    p.content, 
    COALESCE(l.like_count, 0) AS likes, 
    COALESCE(d.dislike_count, 0) AS dislikes
FROM 
    posts p 
JOIN 
    users u ON p.author_id = u.id
LEFT JOIN 
    (SELECT post_id, COUNT(*) AS like_count FROM likes GROUP BY post_id) l ON p.id = l.post_id 
LEFT JOIN 
    (SELECT post_id, COUNT(*) AS dislike_count FROM dislikes GROUP BY post_id) d ON p.id = d.post_id 
LEFT JOIN 
    (SELECT post_id, COUNT(*) AS comment_count FROM comments GROUP BY post_id) c ON p.id = c.post_id
ORDER BY 
    %s %s;
`

const LikedPostsQuery = `
SELECT 
   p.id,
    p.subcategory_id, 
    p.author AS username,
    u.id,
    u.profile_picture_path AS user_image,
    p.event_time,
    p.title,
    p.content, 
    COALESCE(l.like_count, 0) AS likes, 
    COALESCE(d.dislike_count, 0) AS dislikes
FROM 
    posts p 
JOIN 
    users u ON p.author_id = u.id
LEFT JOIN 
    (SELECT post_id, COUNT(*) AS like_count FROM likes GROUP BY post_id) l ON p.id = l.post_id 
LEFT JOIN 
    (SELECT post_id, COUNT(*) AS dislike_count FROM dislikes GROUP BY post_id) d ON p.id = d.post_id 
LEFT JOIN 
    (SELECT post_id, COUNT(*) AS comment_count FROM comments GROUP BY post_id) c ON p.id = c.post_id
JOIN 
    likes lik ON lik.post_id = p.id AND lik.user_id = ?  
ORDER BY 
    p.event_time DESC;
`

const SinglePostQuery = `
SELECT 
    p.id,
    p.subcategory_id, 
    p.author,
    u.id AS author_id,
    u.profile_picture_path AS user_image,
    p.event_time, 
    p.title, 
    p.content, 
    COALESCE(l.like_count, 0) AS likes, 
    COALESCE(d.dislike_count, 0) AS dislikes
FROM 
    posts p 
JOIN 
    users u ON p.author_id = u.id
LEFT JOIN 
    (SELECT post_id, COUNT(*) AS like_count FROM likes GROUP BY post_id) l ON p.id = l.post_id 
LEFT JOIN 
    (SELECT post_id, COUNT(*) AS dislike_count FROM dislikes GROUP BY post_id) d ON p.id = d.post_id 
WHERE 
    p.id = ?
`

const SearchPostQuery = `
SELECT 
    p.id,
    p.subcategory_id, 
    p.author AS username,
    u.id,
    u.profile_picture_path AS user_image,
    p.event_time,
    p.title,
    p.content, 
    COALESCE(l.like_count, 0) AS likes, 
    COALESCE(d.dislike_count, 0) AS dislikes
FROM 
    posts p 
JOIN 
    users u ON p.author_id = u.id
LEFT JOIN 
    (SELECT post_id, COUNT(*) AS like_count FROM likes GROUP BY post_id) l ON p.id = l.post_id 
LEFT JOIN 
    (SELECT post_id, COUNT(*) AS dislike_count FROM dislikes GROUP BY post_id) d ON p.id = d.post_id 
LEFT JOIN 
    (SELECT post_id, COUNT(*) AS comment_count FROM comments GROUP BY post_id) c ON p.id = c.post_id
WHERE
    p.title LIKE ?
ORDER BY 
    %s %s;
`

const CommentQuery = `
SELECT c.id, c.post_id, c.author_id, u.login, u.profile_picture_path, c.content, c.created_at, 
       COALESCE(l.like_count, 0) AS likes, COALESCE(d.dislike_count, 0) AS dislikes
FROM comments c
JOIN users u ON c.author_id = u.id
LEFT JOIN (
    SELECT comment_id, COUNT(*) AS like_count
    FROM likes
    GROUP BY comment_id
) l ON c.id = l.comment_id
LEFT JOIN (
    SELECT comment_id, COUNT(*) AS dislike_count
    FROM dislikes
    GROUP BY comment_id
) d ON c.id = d.comment_id
WHERE c.post_id = ?`

const ReplyQuery = `SELECT r.id, r.comment_id, r.author_id, u.login, r.content, u.profile_picture_path, r.created_at 
FROM replies r 
JOIN users u ON r.author_id = u.id 
WHERE r.comment_id = ?`

const AttachmentQuery = `
    SELECT 
        id,
        attachment_path
    FROM 
        attachments
    WHERE 
        post_id = ?;
`
