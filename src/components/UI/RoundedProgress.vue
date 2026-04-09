<template>
  <div class="circular-progress">
    <svg :width="size" :height="size" viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
      <!-- Background Circle -->
      <circle
          class="progress-background"
          cx="50"
          cy="50"
          r="45"
          fill="none"
          :stroke="backgroundColor"
          :stroke-width="strokeWidth"
      />
      <!-- Progress Circle -->
      <circle
          class="progress"
          cx="50"
          cy="50"
          r="45"
          fill="none"
          :stroke="progressColor"
          :stroke-width="strokeWidth"
          :stroke-dasharray="circumference"
          :stroke-dashoffset="offset"
          stroke-linecap="round"
          transform="rotate(-90 50 50)"
      />
      <!-- Progress Text -->
      <text
          v-if="showText"
          x="50"
          y="50"
          text-anchor="middle"
          dominant-baseline="central"
          :fill="textColor"
          :font-size="textSize"
      >
        {{ Math.round(value) }}%
      </text>
    </svg>
  </div>
</template>

<script setup>
  import {computed} from "vue";

  const props = defineProps({
    value: {
      type: Number,
      required: true,
      validator: v => v >= 0 && v <= 100, // Ensures value is between 0 and 100
    },
    size: {
      type: Number,
      default: 22, // Size in pixels
    },
    strokeWidth: {
      type: Number,
      default: 10, // Thickness of the progress circle
    },
    backgroundColor: {
      type: String,
      default: "#E5E7EB", // Background circle color
    },
    progressColor: {
      type: String,
      default: "#16A34A", // Progress bar color
    },
    textColor: {
      type: String,
      default: "#000", // Text color
    },
    textSize: {
      type: String,
      default: "10px", // Font size of the text
    },
    showText: {
      type: Boolean,
      default: true, // Show percentage text inside the circle
    },
  })

  const circumference = computed(() => {
    return 2 * Math.PI * 45; // 2πr (r = 45 for 50x50 viewbox)
  })

  const offset = computed(() => {
    return circumference.value * (1 - props.value / 100);
  })
</script>

<style scoped>
.circular-progress {
  display: flex;
  position: relative;
}

.progress-background {
  opacity: 0.9;
}
</style>
