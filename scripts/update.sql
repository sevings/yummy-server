select ek.name, sin((ek.karma_raw + ck.karma_raw) * pi() / 2 / 1000) * 100 as karma, 
    ek.karma_raw + ck.karma_raw as karma_raw
from (
    select name, sum(entry_votes.vote) / 2 as karma_raw
    from entry_votes, entries, users
    where abs(entry_votes.vote) > 0.2 and entry_votes.entry_id = entries.id and entries.author_id = users.id
    group by users.name  
) as ek, (
    select name, sum(comment_votes.vote) / 20 as karma_raw
    from comment_votes, comments, users
    where abs(comment_votes.vote) > 0.2 and comment_votes.comment_id = comments.id and comments.author_id = users.id
    group by users.name  
) as ck
where ek.name = ck.name
order by karma_raw desc 
limit 20;
