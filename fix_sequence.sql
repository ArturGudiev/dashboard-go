-- Fix the tasks_id_seq sequence to continue from the max ID in the table
SELECT setval('tasks_id_seq', (SELECT MAX(id) FROM tasks));

