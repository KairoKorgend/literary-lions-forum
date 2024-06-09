CREATE TABLE IF NOT EXISTS users
(
    id       INTEGER     PRIMARY KEY AUTOINCREMENT,
    email    varchar(50)  not null UNIQUE,
    login    varchar(50)  not null UNIQUE,
    hashed_password varchar(255) not null,
    created DATETIME NOT NULL,
    profile_picture_path VARCHAR(255)
);

UPDATE users
SET profile_picture_path = "default-img.jpg"
WHERE profile_picture_path IS NULL;

CREATE TABLE IF NOT EXISTS sessions
(
    id UUID PRIMARY KEY,
    user_id INTEGER,
    created_at DATETIME NOT NULL,
    expires_at DATETIME NOT NULL,
    FOREIGN KEY(user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS categories (
    id     INTEGER      PRIMARY KEY  AUTOINCREMENT,
    title  VARCHAR(10)  UNIQUE  NOT NULL
);

INSERT OR IGNORE INTO categories (id, title) VALUES
(1, 'fiction'),
(2, 'non-fiction'),
(3, 'special interest'),
(4, 'literary forms');

CREATE TABLE IF NOT EXISTS subcategories (
    id           INTEGER       PRIMARY KEY  AUTOINCREMENT,
    title        VARCHAR(10)   UNIQUE  NOT NULL,
    category_id  INTEGER       NOT NULL,
    icon_path    VARCHAR(255)  NOT NULL,
    FOREIGN KEY(category_id)  REFERENCES categories(id)
);

INSERT OR IGNORE INTO subcategories (title, category_id, icon_path) VALUES
('Fantasy', 1, 'wizard.png'),
('Mystery', 1, 'mystery.png'),
('Romance', 1, 'love-books.png'),
('Horror',  1, 'scream.png'),
('Biography', 2, 'biography.png'),
('Self-Help', 2, 'self-confidence.png'),
('History', 2, 'parchment.png'),
('Science', 2, 'science.png'),
('Cooking', 3, 'cooking.png'),
('Sports', 3, 'sports.png'),
('Politics', 3, 'politics.png'),
('Religion', 3, 'religion.png'),
('Poetry', 4, 'poetry.png'),
('Short Stories', 4, 'storybook.png'),
('Comics', 4, 'comic.png');

CREATE TABLE IF NOT EXISTS posts (
    id              INTEGER        PRIMARY KEY  AUTOINCREMENT,
    subcategory_id  INTEGER,
    author            VARCHAR(50)     NOT NULL,
    author_id         INTEGER         NOT NULL,
    event_time      TIMESTAMP       NOT NULL,
    title           VARCHAR(50)     NOT NULL,
    content         VARCHAR(5000)   NOT NULL,
    FOREIGN KEY(subcategory_id)  REFERENCES subcategories(id),
    FOREIGN KEY(author_id)         REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS likes (
    id         INTEGER           PRIMARY KEY  AUTOINCREMENT,
    post_id    INTEGER,
    comment_id INTEGER,
    user_id    INTEGER,
    FOREIGN KEY(post_id)         REFERENCES posts(id),
    FOREIGN KEY(comment_id)      REFERENCES comments(id),
    FOREIGN KEY(user_id)         REFERENCES users(id),
    CHECK (post_id IS NOT NULL OR comment_id IS NOT NULL)
);

CREATE TABLE IF NOT EXISTS dislikes (
    id         INTEGER           PRIMARY KEY  AUTOINCREMENT,
    post_id    INTEGER,
    comment_id INTEGER,
    user_id    INTEGER,
    FOREIGN KEY(post_id)         REFERENCES posts(id),
    FOREIGN KEY(comment_id)      REFERENCES comments(id),
    FOREIGN KEY(user_id)         REFERENCES users(id),
    CHECK (post_id IS NOT NULL OR comment_id IS NOT NULL)
);


CREATE TABLE IF NOT EXISTS replies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    comment_id         INTEGER NOT NULL,
    author_id         INTEGER NOT NULL,
    author_name       VARCHAR(255) NOT NULL,
    content         TEXT NOT NULL,
    created_at      DATETIME DEFAULT    CURRENT_TIMESTAMP,
    FOREIGN KEY(comment_id) REFERENCES comments(id),
    FOREIGN KEY(author_id) REFERENCES users(id)
);


CREATE TABLE IF NOT EXISTS comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id         INTEGER NOT NULL,
    author_id         INTEGER NOT NULL,
    author_name       VARCHAR(255) NOT NULL,
    content         TEXT NOT NULL,
    created_at      DATETIME DEFAULT    CURRENT_TIMESTAMP,
    FOREIGN KEY(post_id) REFERENCES posts(id),
    FOREIGN KEY(author_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS attachments (
    id                    INTEGER       PRIMARY KEY AUTOINCREMENT,
    post_id               INTEGER       NOT NULL,
    attachment_path       VARCHAR(255)  NOT NULL,
    FOREIGN KEY(post_id)  REFERENCES posts(id)
);