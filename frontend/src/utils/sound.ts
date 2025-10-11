import dingSound from '../sounds/ding.mp3';
import { LogError } from "@/api/log";

const audioCache: Record<string, HTMLAudioElement> = {};

export enum SoundEffect {
  TaskComplete = 'task-complete',
}

const soundFiles: Record<SoundEffect, string> = {
  [SoundEffect.TaskComplete]: dingSound,
};

export const preloadSounds = (): void => {
  Object.entries(soundFiles).forEach(([effect, url]) => {
    const audio = new Audio(url);
    audio.preload = 'auto';
    audioCache[effect] = audio;
    
    audio.load();
  });
};

export const playSound = (effect: SoundEffect): void => {
  const audio = audioCache[effect] || new Audio(soundFiles[effect]);
  
  if (!audioCache[effect]) {
    audioCache[effect] = audio;
  }
  
  audio.pause();
  audio.currentTime = 0;
  
  audio.play().catch(error => {
    LogError(`Failed to play sound [${effect}]: ${error}`, window.location.pathname)
  });
};
