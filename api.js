const express = require('express');
const fs = require('fs');
const path = require('path');
const NodeCache = require('node-cache');

const app = express();
const PORT = process.env.PORT || 3000;

// Initialize cache with 5-minute standard TTL
const cache = new NodeCache({ stdTTL: 300, checkperiod: 60 });

let exercisesData;

// Load exercises data
function loadExercises() {
  try {
    const data = fs.readFileSync(path.join(__dirname, 'exercises.json'), 'utf8');
    exercisesData = JSON.parse(data);
    console.log(`Loaded ${exercisesData.exercises.length} exercises`);
  } catch (err) {
    console.error('Failed to load exercises:', err);
    process.exit(1);
  }
}

// Generate cache key from parameters
function generateCacheKey(muscles, equipment, category) {
  const key = muscles.join(',') + '|' + equipment;
  return category ? `${key}|${category}` : key;
}

// Filter exercises based on parameters
function filterExercises(muscles, equipment, category) {
  // Normalize inputs to lowercase
  const normalizedMuscles = muscles.map(m => m.toLowerCase());
  const normalizedEquipment = equipment.toLowerCase() || 'none';

  return exercisesData.exercises.filter(exercise => {
    // Check equipment filter
    const equipmentMatch = exercise.equipment.some(
      eq => eq.toLowerCase() === normalizedEquipment
    );
    if (!equipmentMatch) return false;

    // Check muscle group filter - must match at least one required muscle
    const muscleMatch = normalizedMuscles.some(
      muscle => exercise.muscle_group.toLowerCase() === muscle
    );
    if (!muscleMatch) return false;

    // Check category filter if provided
    if (category && exercise.category.toLowerCase() !== category.toLowerCase()) {
      return false;
    }

    return true;
  });
}

// Routes
app.get('/health', (req, res) => {
  res.json({
    status: 'ok',
    timestamp: new Date().toISOString(),
  });
});

app.get('/exercises', (req, res) => {
  try {
    // Parse and validate required parameters
    const muscleParam = req.query.muscle_group;
    const equipment = req.query.equipment || 'none';
    const category = req.query.category;

    if (!muscleParam) {
      return res.status(400).json({
        error: 'muscle_group parameter is required',
      });
    }

    // Split multiple muscle groups (comma or space separated)
    const muscles = muscleParam
      .split(',')
      .map(m => m.trim())
      .filter(m => m);

    // Generate cache key
    const cacheKey = generateCacheKey(muscles, equipment, category);

    // Check cache first
    let cachedResult = cache.get(cacheKey);
    if (cachedResult) {
      return res.json({
        data: cachedResult,
        count: cachedResult.length,
        cached: true,
        timestamp: new Date().toISOString(),
      });
    }

    // Filter exercises
    const results = filterExercises(muscles, equipment, category);

    // Store in cache
    cache.set(cacheKey, results);

    res.json({
      data: results,
      count: results.length,
      cached: false,
      timestamp: new Date().toISOString(),
    });
  } catch (err) {
    console.error('Error:', err);
    res.status(500).json({
      error: 'Internal server error',
    });
  }
});

app.get('/stats', (req, res) => {
  res.json({
    categories: exercisesData.categories,
    equipment: exercisesData.equipment,
    muscle_groups: exercisesData.muscle_groups,
    total_exercises: exercisesData.exercises.length,
  });
});

// Load data and start server
loadExercises();

app.listen(PORT, () => {
  console.log(`API server running on port ${PORT}`);
});
