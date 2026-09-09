import axios, { AxiosInstance } from 'axios';
import type { Probe, Target, Monitor, Measurement, TargetStats } from '../types';

class ApiService {
  private client: AxiosInstance;
  private baseURL: string;

  constructor() {
    this.baseURL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';
    this.client = axios.create({
      baseURL: this.baseURL,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Request interceptor
    this.client.interceptors.request.use(
      (config) => {
        const token = localStorage.getItem('api_token');
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => Promise.reject(error)
    );

    // Response interceptor
    this.client.interceptors.response.use(
      (response) => response,
      (error) => {
        console.error('API Error:', error.response?.data || error.message);
        return Promise.reject(error);
      }
    );
  }

  // Health check
  async healthCheck(): Promise<{ status: string; time: string }> {
    const response = await axios.get(`${this.baseURL.replace('/api/v1', '')}/health`);
    return response.data;
  }

  // Probes
  async getProbes(): Promise<Probe[]> {
    const response = await this.client.get<Probe[]>('/probes');
    return response.data;
  }

  async getProbe(id: string): Promise<Probe> {
    const response = await this.client.get<Probe>(`/probes/${id}`);
    return response.data;
  }

  async createProbe(probe: Partial<Probe>): Promise<Probe> {
    const response = await this.client.post<Probe>('/probes', probe);
    return response.data;
  }

  async probeHeartbeat(id: string): Promise<void> {
    await this.client.post(`/probes/${id}/heartbeat`);
  }

  // Targets
  async getTargets(enabledOnly = false): Promise<Target[]> {
    const response = await this.client.get<Target[]>('/targets', {
      params: { enabled: enabledOnly },
    });
    return response.data;
  }

  async getTarget(id: string): Promise<Target> {
    const response = await this.client.get<Target>(`/targets/${id}`);
    return response.data;
  }

  async createTarget(target: Partial<Target>): Promise<Target> {
    const response = await this.client.post<Target>('/targets', target);
    return response.data;
  }

  async getTargetStats(id: string): Promise<TargetStats> {
    const response = await this.client.get<TargetStats>(`/targets/${id}/stats`);
    return response.data;
  }

  async getTargetMeasurements(id: string): Promise<Measurement[]> {
    const response = await this.client.get<Measurement[]>(`/targets/${id}/measurements`);
    return response.data;
  }

  // Monitors
  async getMonitors(enabledOnly = false): Promise<Monitor[]> {
    const response = await this.client.get<Monitor[]>('/monitors', {
      params: { enabled: enabledOnly },
    });
    return response.data;
  }

  async getMonitor(id: string): Promise<Monitor> {
    const response = await this.client.get<Monitor>(`/monitors/${id}`);
    return response.data;
  }

  async createMonitor(monitor: Partial<Monitor>): Promise<Monitor> {
    const response = await this.client.post<Monitor>('/monitors', monitor);
    return response.data;
  }

  // Measurements
  async submitMeasurement(measurement: Partial<Measurement>): Promise<void> {
    await this.client.post('/measurements', measurement);
  }

  // Dashboard stats (aggregated from targets)
  async getDashboardStats(): Promise<any> {
    const targets = await this.getTargets(true);
    const probes = await this.getProbes();

    const stats = {
      total_targets: targets.length,
      targets_up: 0,
      targets_down: 0,
      targets_degraded: 0,
      total_probes: probes.length,
      active_probes: probes.filter(p => p.status === 'active').length,
      active_alerts: 0,
      active_incidents: 0,
    };

    // Get stats for each target to determine status
    const statsPromises = targets.map(t => this.getTargetStats(t.id).catch(() => null));
    const targetStats = await Promise.all(statsPromises);

    targetStats.forEach(stat => {
      if (stat) {
        const availability = stat.availability ?? 0;
        if (availability >= 99) {
          stats.targets_up++;
        } else if (availability >= 50) {
          stats.targets_degraded++;
        } else {
          stats.targets_down++;
        }
      }
    });

    return stats;
  }
}

export const apiService = new ApiService();
export default apiService;
