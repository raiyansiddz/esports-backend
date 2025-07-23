package services

import (
        "encoding/json"
        "fmt"
        "log"
        "math/rand"
        "time"

        "esports-fantasy-backend/config"
        "esports-fantasy-backend/internal/models"
        "esports-fantasy-backend/internal/repository"

        "github.com/go-redis/redis/v8"
        "github.com/gorilla/websocket"
)

type MatchSimulationService struct {
        cfg                *config.Config
        matchRepo          repository.MatchRepository
        playerRepo         repository.PlayerRepository
        scoringService     *ScoringService
        leaderboardService *LeaderboardService
        rdb                *redis.Client
        activeSimulations  map[string]*MatchSimulation
}

type MatchSimulation struct {
        MatchID     string
        Match       *models.Match
        Players     []*models.Player
        Events      []MatchEvent
        IsActive    bool
        StartTime   time.Time
        CurrentTime time.Time
        Clients     map[string]*websocket.Conn
}

type MatchEvent struct {
        ID          string                 `json:"id"`
        MatchID     string                 `json:"match_id"`
        PlayerID    string                 `json:"player_id"`
        PlayerName  string                 `json:"player_name"`
        EventType   string                 `json:"event_type"` // KILL, KNOCKOUT, REVIVE, DEATH, MVP
        Points      float64                `json:"points"`
        Timestamp   time.Time              `json:"timestamp"`
        Description string                 `json:"description"`
        Metadata    map[string]interface{} `json:"metadata"`
}

type LiveMatchUpdate struct {
        Type      string      `json:"type"`
        MatchID   string      `json:"match_id"`
        Timestamp time.Time   `json:"timestamp"`
        Data      interface{} `json:"data"`
}

type MatchEventUpdate struct {
        Event           MatchEvent             `json:"event"`
        PlayerStats     map[string]PlayerStat  `json:"player_stats"`
        Leaderboard     []LeaderboardEntry     `json:"leaderboard"`
        MatchTimer      string                 `json:"match_timer"`
        PlayersAlive    int                    `json:"players_alive"`
}

type PlayerStat struct {
        PlayerID     string  `json:"player_id"`
        PlayerName   string  `json:"player_name"`
        Kills        int     `json:"kills"`
        Knockouts    int     `json:"knockouts"`
        Revives      int     `json:"revives"`
        Points       float64 `json:"points"`
        IsAlive      bool    `json:"is_alive"`
        IsMVP        bool    `json:"is_mvp"`
        SurvivalTime int     `json:"survival_time_minutes"`
}

type LeaderboardEntry struct {
        Rank       int     `json:"rank"`
        TeamName   string  `json:"team_name"`
        UserName   string  `json:"user_name"`
        Points     float64 `json:"points"`
        Movement   string  `json:"movement"` // UP, DOWN, SAME
}

func NewMatchSimulationService(cfg *config.Config, matchRepo repository.MatchRepository, playerRepo repository.PlayerRepository, scoringService *ScoringService, leaderboardService *LeaderboardService, rdb *redis.Client) *MatchSimulationService {
        return &MatchSimulationService{
                cfg:                cfg,
                matchRepo:          matchRepo,
                playerRepo:         playerRepo,
                scoringService:     scoringService,
                leaderboardService: leaderboardService,
                rdb:                rdb,
                activeSimulations:  make(map[string]*MatchSimulation),
        }
}

func (s *MatchSimulationService) StartMatchSimulation(matchID string) error {
        if !s.cfg.MatchSimulationEnabled {
                return fmt.Errorf("match simulation is disabled")
        }

        // Check if simulation already active
        if _, exists := s.activeSimulations[matchID]; exists {
                return fmt.Errorf("simulation already active for match %s", matchID)
        }

        // Get match details
        match, err := s.matchRepo.GetByID(matchID)
        if err != nil {
                return fmt.Errorf("failed to get match: %w", err)
        }

        if match.Status != "LIVE" {
                return fmt.Errorf("match is not live")
        }

        // Get players for the match
        players, err := s.playerRepo.GetPlayersByMatchID(matchID)
        if err != nil {
                return fmt.Errorf("failed to get players: %w", err)
        }

        // Create simulation
        simulation := &MatchSimulation{
                MatchID:     matchID,
                Match:       match,
                Players:     players,
                Events:      []MatchEvent{},
                IsActive:    true,
                StartTime:   time.Now(),
                CurrentTime: time.Now(),
                Clients:     make(map[string]*websocket.Conn),
        }

        s.activeSimulations[matchID] = simulation

        // Start simulation goroutine
        go s.runMatchSimulation(simulation)

        log.Printf("🎮 Match simulation started for: %s (%s)", match.Name, matchID[:8])
        return nil
}

func (s *MatchSimulationService) StopMatchSimulation(matchID string) error {
        simulation, exists := s.activeSimulations[matchID]
        if !exists {
                return fmt.Errorf("no active simulation for match %s", matchID)
        }

        simulation.IsActive = false
        delete(s.activeSimulations, matchID)

        // Close all WebSocket connections
        for _, conn := range simulation.Clients {
                conn.Close()
        }

        log.Printf("⏹️ Match simulation stopped for: %s", matchID[:8])
        return nil
}

func (s *MatchSimulationService) runMatchSimulation(simulation *MatchSimulation) {
        defer func() {
                if r := recover(); r != nil {
                        log.Printf("❌ Match simulation panic for %s: %v", simulation.MatchID, r)
                }
        }()

        matchDuration := 25 * time.Minute // Typical BGMI match duration
        endTime := simulation.StartTime.Add(matchDuration)
        
        // Initialize player stats
        playerStats := make(map[string]PlayerStat)
        for _, player := range simulation.Players {
                playerStats[player.ID] = PlayerStat{
                        PlayerID:     player.ID,
                        PlayerName:   player.Name,
                        Kills:        0,
                        Knockouts:    0,
                        Revives:      0,
                        Points:       0,
                        IsAlive:      true,
                        IsMVP:        false,
                        SurvivalTime: 0,
                }
        }

        playersAlive := len(simulation.Players)
        eventCounter := 0

        log.Printf("🔴 LIVE SIMULATION: %s - %d players", simulation.Match.Name, playersAlive)

        for simulation.IsActive && time.Now().Before(endTime) {
                // Random event generation every 10-30 seconds
                eventInterval := time.Duration(10+rand.Intn(20)) * time.Second
                time.Sleep(eventInterval)

                if !simulation.IsActive {
                        break
                }

                // Generate random event
                event := s.generateRandomEvent(simulation, playerStats, playersAlive)
                if event != nil {
                        eventCounter++
                        simulation.Events = append(simulation.Events, *event)

                        // Update player stats based on event
                        s.updatePlayerStats(event, playerStats)

                        // Update survival time for all alive players
                        survivedMinutes := int(time.Since(simulation.StartTime).Minutes())
                        for id, stat := range playerStats {
                                if stat.IsAlive {
                                        stat.SurvivalTime = survivedMinutes
                                        playerStats[id] = stat
                                }
                        }

                        // Calculate points
                        s.calculateEventPoints(event, playerStats)

                        // Check if player died
                        if event.EventType == "DEATH" {
                                playersAlive--
                                stat := playerStats[event.PlayerID]
                                stat.IsAlive = false
                                playerStats[event.PlayerID] = stat
                        }

                        // Get updated leaderboard
                        leaderboard := s.getSimulatedLeaderboard(simulation.MatchID)

                        // Broadcast update
                        update := MatchEventUpdate{
                                Event:        *event,
                                PlayerStats:  playerStats,
                                Leaderboard:  leaderboard,
                                MatchTimer:   s.formatMatchTimer(simulation.StartTime),
                                PlayersAlive: playersAlive,
                        }

                        s.broadcastMatchUpdate(simulation, "match_event", update)

                        // Log significant events
                        if event.EventType == "KILL" || event.EventType == "DEATH" {
                                log.Printf("🎯 %s: %s (%s) - Players alive: %d", 
                                        event.EventType, event.PlayerName, event.Description, playersAlive)
                        }

                        // End match if only few players left
                        if playersAlive <= 3 {
                                break
                        }
                }
        }

        // Match ended - determine MVP and finalize
        s.finalizeMatchSimulation(simulation, playerStats)
}

func (s *MatchSimulationService) generateRandomEvent(simulation *MatchSimulation, playerStats map[string]PlayerStat, playersAlive int) *MatchEvent {
        // Get alive players
        var alivePlayers []models.Player
        for _, player := range simulation.Players {
                if stat, exists := playerStats[player.ID]; exists && stat.IsAlive {
                        alivePlayers = append(alivePlayers, *player)
                }
        }

        if len(alivePlayers) == 0 {
                return nil
        }

        // Random player
        randomPlayer := alivePlayers[rand.Intn(len(alivePlayers))]
        
        // Event probabilities
        eventTypes := []string{"KILL", "KNOCKOUT", "REVIVE", "DEATH"}
        weights := []int{30, 40, 20, 10} // Higher chance for kills and knockouts

        // Adjust death probability based on match progress
        matchProgress := time.Since(simulation.StartTime).Minutes() / 25.0
        if matchProgress > 0.7 { // Late game - more deaths
                weights[3] = 25
        }

        // Select random event type
        eventType := s.weightedRandomSelect(eventTypes, weights)

        // Create event
        event := &MatchEvent{
                ID:        fmt.Sprintf("event_%s_%d", simulation.MatchID, len(simulation.Events)+1),
                MatchID:   simulation.MatchID,
                PlayerID:  randomPlayer.ID,
                PlayerName: randomPlayer.Name,
                EventType: eventType,
                Timestamp: time.Now(),
                Metadata:  make(map[string]interface{}),
        }

        // Set event-specific data
        switch eventType {
        case "KILL":
                event.Description = fmt.Sprintf("%s eliminated an opponent!", randomPlayer.Name)
                event.Points = 10
                event.Metadata["weapon"] = s.getRandomWeapon()
                
        case "KNOCKOUT":
                event.Description = fmt.Sprintf("%s knocked down an enemy!", randomPlayer.Name)
                event.Points = 6
                
        case "REVIVE":
                event.Description = fmt.Sprintf("%s revived a teammate!", randomPlayer.Name)
                event.Points = 5
                
        case "DEATH":
                // Only if match is progressed enough or player has been active
                if matchProgress > 0.3 {
                        event.Description = fmt.Sprintf("%s was eliminated!", randomPlayer.Name)
                        event.Points = 0
                        event.Metadata["rank"] = playersAlive
                } else {
                        return nil // Don't kill players too early
                }
        }

        return event
}

func (s *MatchSimulationService) updatePlayerStats(event *MatchEvent, playerStats map[string]PlayerStat) {
        stat := playerStats[event.PlayerID]
        
        switch event.EventType {
        case "KILL":
                stat.Kills++
        case "KNOCKOUT":
                stat.Knockouts++
        case "REVIVE":
                stat.Revives++
        case "DEATH":
                stat.IsAlive = false
        }
        
        playerStats[event.PlayerID] = stat
}

func (s *MatchSimulationService) calculateEventPoints(event *MatchEvent, playerStats map[string]PlayerStat) {
        stat := playerStats[event.PlayerID]
        stat.Points += event.Points
        playerStats[event.PlayerID] = stat
}

func (s *MatchSimulationService) finalizeMatchSimulation(simulation *MatchSimulation, playerStats map[string]PlayerStat) {
        // Determine MVP based on points
        var mvpPlayerID string
        maxPoints := 0.0
        
        for playerID, stat := range playerStats {
                if stat.Points > maxPoints {
                        maxPoints = stat.Points
                        mvpPlayerID = playerID
                }
        }

        // Set MVP
        if mvpPlayerID != "" {
                stat := playerStats[mvpPlayerID]
                stat.IsMVP = true
                stat.Points += 20 // MVP bonus
                playerStats[mvpPlayerID] = stat
                
                log.Printf("👑 MVP: %s with %.1f points", stat.PlayerName, stat.Points)
        }

        // Save final stats to database
        s.saveFinalStats(simulation, playerStats)

        // Update match status
        simulation.Match.Status = "COMPLETED"
        s.matchRepo.Update(simulation.Match)

        // Final leaderboard update
        finalLeaderboard := s.getSimulatedLeaderboard(simulation.MatchID)
        finalUpdate := LiveMatchUpdate{
                Type:      "match_completed",
                MatchID:   simulation.MatchID,
                Timestamp: time.Now(),
                Data: map[string]interface{}{
                        "final_stats":   playerStats,
                        "leaderboard":   finalLeaderboard,
                        "mvp_player":    mvpPlayerID,
                        "total_events":  len(simulation.Events),
                        "match_duration": time.Since(simulation.StartTime).Minutes(),
                },
        }

        s.broadcastMatchUpdate(simulation, "match_completed", finalUpdate.Data)

        log.Printf("🏁 Match simulation completed: %s", simulation.Match.Name)
        
        // Clean up
        s.StopMatchSimulation(simulation.MatchID)
}

func (s *MatchSimulationService) saveFinalStats(simulation *MatchSimulation, playerStats map[string]PlayerStat) {
        for _, stat := range playerStats {
                matchStat := &models.PlayerMatchStats{
                        ID:                  fmt.Sprintf("sim_%s_%s", simulation.MatchID, stat.PlayerID),
                        PlayerID:            stat.PlayerID,
                        MatchID:             simulation.MatchID,
                        Kills:               stat.Kills,
                        Knockouts:           stat.Knockouts,
                        Revives:             stat.Revives,
                        SurvivalTimeMinutes: stat.SurvivalTime,
                        Points:              stat.Points,
                        IsMVP:               stat.IsMVP,
                        CreatedAt:           time.Now(),
                        UpdatedAt:           time.Now(),
                }

                // This would be saved to database in production
                log.Printf("📊 Final stats for %s: %d kills, %d knockouts, %.1f points", 
                        stat.PlayerName, stat.Kills, stat.Knockouts, stat.Points)
        }
}

func (s *MatchSimulationService) getSimulatedLeaderboard(matchID string) []LeaderboardEntry {
        // In production, this would get real leaderboard from LeaderboardService
        // For simulation, return mock data
        return []LeaderboardEntry{
                {Rank: 1, TeamName: "Pro Gamers", UserName: "User1", Points: 245.5, Movement: "UP"},
                {Rank: 2, TeamName: "Elite Squad", UserName: "User2", Points: 220.0, Movement: "DOWN"},
                {Rank: 3, TeamName: "Winners", UserName: "User3", Points: 195.5, Movement: "SAME"},
        }
}

func (s *MatchSimulationService) broadcastMatchUpdate(simulation *MatchSimulation, eventType string, data interface{}) {
        update := LiveMatchUpdate{
                Type:      eventType,
                MatchID:   simulation.MatchID,
                Timestamp: time.Now(),
                Data:      data,
        }

        message, err := json.Marshal(update)
        if err != nil {
                log.Printf("❌ Error marshaling update: %v", err)
                return
        }

        // Broadcast to all connected clients
        for clientID, conn := range simulation.Clients {
                if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
                        log.Printf("❌ Error sending message to client %s: %v", clientID, err)
                        conn.Close()
                        delete(simulation.Clients, clientID)
                }
        }

        // Cache in Redis for clients that reconnect
        cacheKey := fmt.Sprintf("match_updates:%s", simulation.MatchID)
        s.rdb.LPush(ctx, cacheKey, message)
        s.rdb.LTrim(ctx, cacheKey, 0, 100) // Keep last 100 updates
        s.rdb.Expire(ctx, cacheKey, 2*time.Hour)
}

func (s *MatchSimulationService) AddWebSocketClient(matchID, clientID string, conn *websocket.Conn) error {
        simulation, exists := s.activeSimulations[matchID]
        if !exists {
                return fmt.Errorf("no active simulation for match %s", matchID)
        }

        simulation.Clients[clientID] = conn
        log.Printf("🔌 Client connected to match %s: %s", matchID[:8], clientID)

        // Send recent updates to new client
        s.sendRecentUpdates(matchID, conn)
        
        return nil
}

func (s *MatchSimulationService) RemoveWebSocketClient(matchID, clientID string) {
        if simulation, exists := s.activeSimulations[matchID]; exists {
                if conn, exists := simulation.Clients[clientID]; exists {
                        conn.Close()
                        delete(simulation.Clients, clientID)
                        log.Printf("🔌 Client disconnected from match %s: %s", matchID[:8], clientID)
                }
        }
}

func (s *MatchSimulationService) sendRecentUpdates(matchID string, conn *websocket.Conn) {
        cacheKey := fmt.Sprintf("match_updates:%s", matchID)
        updates, err := s.rdb.LRange(ctx, cacheKey, 0, 10).Result()
        if err != nil {
                return
        }

        for i := len(updates) - 1; i >= 0; i-- {
                conn.WriteMessage(websocket.TextMessage, []byte(updates[i]))
        }
}

// Helper functions
func (s *MatchSimulationService) weightedRandomSelect(items []string, weights []int) string {
        total := 0
        for _, w := range weights {
                total += w
        }

        r := rand.Intn(total)
        for i, w := range weights {
                r -= w
                if r < 0 {
                        return items[i]
                }
        }
        return items[0]
}

func (s *MatchSimulationService) getRandomWeapon() string {
        weapons := []string{"AKM", "M416", "SCAR-L", "UMP45", "Vector", "AWM", "Kar98k", "M24"}
        return weapons[rand.Intn(len(weapons))]
}

func (s *MatchSimulationService) formatMatchTimer(startTime time.Time) string {
        elapsed := time.Since(startTime)
        minutes := int(elapsed.Minutes())
        seconds := int(elapsed.Seconds()) % 60
        return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func (s *MatchSimulationService) GetActiveSimulations() map[string]*MatchSimulation {
        return s.activeSimulations
}

func (s *MatchSimulationService) GetMatchEvents(matchID string) ([]MatchEvent, error) {
        simulation, exists := s.activeSimulations[matchID]
        if !exists {
                return nil, fmt.Errorf("no active simulation for match %s", matchID)
        }
        return simulation.Events, nil
}