CREATE TABLE IF NOT EXISTS refresh_tokens(

    id SERIAL PRIMARY KEY,

    -- the actual token we are going to validate against
    token TEXT UNIQUE NOT NULL,

    -- the id of the user to whom this token belongs, 
    -- references users(id) is our foreign key
    -- the ON DELETE CASCADE will delete the refresh token when the user is deleted
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()

);