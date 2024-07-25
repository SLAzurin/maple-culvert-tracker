select maple_character_name
from character_culvert_scores
inner join characters on characters.id = character_culvert_scores.character_id
where
    culvert_date = '2024-07-21'
and score = 0;

-- select max scores of only active members
select characters.id, maple_character_name, count(character_culvert_scores.id) as num_of_weeks_recorded_to_date, max(character_culvert_scores.score) as max_recorded_score
from character_culvert_scores
inner join ( -- this subquery filters only current members not older members
    select character_culvert_scores.character_id as id 
    from character_culvert_scores
    where culvert_date = '2024-07-21'
) as scoped_current_characters on scoped_current_characters.id = character_culvert_scores.character_id
inner join characters on characters.id = character_culvert_scores.character_id
group by characters.id, maple_character_name
order by max_recorded_score desc;

-- select count # of weeks and # of non-sandbagged runs
select base_all.id, base_all.maple_character_name, base_all.num_of_weeks_recorded_to_date, COALESCE(sandbagged.num_of_weeks_sandbagged, 0) as num_of_weeks_sandbagged from 
(select characters.id as id, maple_character_name, count(character_culvert_scores.id) as num_of_weeks_recorded_to_date
from character_culvert_scores
inner join ( -- this subquery filters only current members not older members
    select character_culvert_scores.character_id as id 
    from character_culvert_scores
    where culvert_date = '2024-07-21'
) as scoped_current_characters on scoped_current_characters.id = character_culvert_scores.character_id
inner join characters on characters.id = character_culvert_scores.character_id
group by characters.id, maple_character_name) as base_all

left join (select characters.id, maple_character_name, count(character_culvert_scores.id) as num_of_weeks_sandbagged
from character_culvert_scores
inner join ( -- this subquery filters only current members not older members
    select character_culvert_scores.character_id as id 
    from character_culvert_scores
    where culvert_date = '2024-07-21'
) as scoped_current_characters on scoped_current_characters.id = character_culvert_scores.character_id
inner join characters on characters.id = character_culvert_scores.character_id
where character_culvert_scores.score = 0
group by characters.id, maple_character_name) as sandbagged on sandbagged.id = base_all.id
order by num_of_weeks_sandbagged desc;

/*
Goal: Find all characters that:
within the past 12 weeks,
number of weeks sandbagged /12
average of non-zero score within those 12 weeks,
participation ratio /12 in %

also, sandbagged scores are scores that fall below 70% of the previous week's score

*/