// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'profile.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

T _$identity<T>(T value) => value;

final _privateConstructorUsedError = UnsupportedError(
    'It seems like you constructed your class using `MyClass._()`. This constructor is only meant to be used by freezed and you are not supposed to need it nor use it.\nPlease check the documentation here for more information: https://github.com/rrousselGit/freezed#adding-getters-and-methods-to-our-models');

MemberProfile _$MemberProfileFromJson(Map<String, dynamic> json) {
  return _MemberProfile.fromJson(json);
}

/// @nodoc
mixin _$MemberProfile {
  ProfileUser get user => throw _privateConstructorUsedError;
  UserStats get stats => throw _privateConstructorUsedError;
  List<ProfileAchievement> get achievements =>
      throw _privateConstructorUsedError;
  String get legend => throw _privateConstructorUsedError;
  @JsonKey(name: 'season_history')
  List<SeasonCardDto> get seasonHistory => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $MemberProfileCopyWith<MemberProfile> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $MemberProfileCopyWith<$Res> {
  factory $MemberProfileCopyWith(
          MemberProfile value, $Res Function(MemberProfile) then) =
      _$MemberProfileCopyWithImpl<$Res, MemberProfile>;
  @useResult
  $Res call(
      {ProfileUser user,
      UserStats stats,
      List<ProfileAchievement> achievements,
      String legend,
      @JsonKey(name: 'season_history') List<SeasonCardDto> seasonHistory});

  $ProfileUserCopyWith<$Res> get user;
  $UserStatsCopyWith<$Res> get stats;
}

/// @nodoc
class _$MemberProfileCopyWithImpl<$Res, $Val extends MemberProfile>
    implements $MemberProfileCopyWith<$Res> {
  _$MemberProfileCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? user = null,
    Object? stats = null,
    Object? achievements = null,
    Object? legend = null,
    Object? seasonHistory = null,
  }) {
    return _then(_value.copyWith(
      user: null == user
          ? _value.user
          : user // ignore: cast_nullable_to_non_nullable
              as ProfileUser,
      stats: null == stats
          ? _value.stats
          : stats // ignore: cast_nullable_to_non_nullable
              as UserStats,
      achievements: null == achievements
          ? _value.achievements
          : achievements // ignore: cast_nullable_to_non_nullable
              as List<ProfileAchievement>,
      legend: null == legend
          ? _value.legend
          : legend // ignore: cast_nullable_to_non_nullable
              as String,
      seasonHistory: null == seasonHistory
          ? _value.seasonHistory
          : seasonHistory // ignore: cast_nullable_to_non_nullable
              as List<SeasonCardDto>,
    ) as $Val);
  }

  @override
  @pragma('vm:prefer-inline')
  $ProfileUserCopyWith<$Res> get user {
    return $ProfileUserCopyWith<$Res>(_value.user, (value) {
      return _then(_value.copyWith(user: value) as $Val);
    });
  }

  @override
  @pragma('vm:prefer-inline')
  $UserStatsCopyWith<$Res> get stats {
    return $UserStatsCopyWith<$Res>(_value.stats, (value) {
      return _then(_value.copyWith(stats: value) as $Val);
    });
  }
}

/// @nodoc
abstract class _$$MemberProfileImplCopyWith<$Res>
    implements $MemberProfileCopyWith<$Res> {
  factory _$$MemberProfileImplCopyWith(
          _$MemberProfileImpl value, $Res Function(_$MemberProfileImpl) then) =
      __$$MemberProfileImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {ProfileUser user,
      UserStats stats,
      List<ProfileAchievement> achievements,
      String legend,
      @JsonKey(name: 'season_history') List<SeasonCardDto> seasonHistory});

  @override
  $ProfileUserCopyWith<$Res> get user;
  @override
  $UserStatsCopyWith<$Res> get stats;
}

/// @nodoc
class __$$MemberProfileImplCopyWithImpl<$Res>
    extends _$MemberProfileCopyWithImpl<$Res, _$MemberProfileImpl>
    implements _$$MemberProfileImplCopyWith<$Res> {
  __$$MemberProfileImplCopyWithImpl(
      _$MemberProfileImpl _value, $Res Function(_$MemberProfileImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? user = null,
    Object? stats = null,
    Object? achievements = null,
    Object? legend = null,
    Object? seasonHistory = null,
  }) {
    return _then(_$MemberProfileImpl(
      user: null == user
          ? _value.user
          : user // ignore: cast_nullable_to_non_nullable
              as ProfileUser,
      stats: null == stats
          ? _value.stats
          : stats // ignore: cast_nullable_to_non_nullable
              as UserStats,
      achievements: null == achievements
          ? _value._achievements
          : achievements // ignore: cast_nullable_to_non_nullable
              as List<ProfileAchievement>,
      legend: null == legend
          ? _value.legend
          : legend // ignore: cast_nullable_to_non_nullable
              as String,
      seasonHistory: null == seasonHistory
          ? _value._seasonHistory
          : seasonHistory // ignore: cast_nullable_to_non_nullable
              as List<SeasonCardDto>,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$MemberProfileImpl implements _MemberProfile {
  const _$MemberProfileImpl(
      {required this.user,
      required this.stats,
      required final List<ProfileAchievement> achievements,
      required this.legend,
      @JsonKey(name: 'season_history')
      required final List<SeasonCardDto> seasonHistory})
      : _achievements = achievements,
        _seasonHistory = seasonHistory;

  factory _$MemberProfileImpl.fromJson(Map<String, dynamic> json) =>
      _$$MemberProfileImplFromJson(json);

  @override
  final ProfileUser user;
  @override
  final UserStats stats;
  final List<ProfileAchievement> _achievements;
  @override
  List<ProfileAchievement> get achievements {
    if (_achievements is EqualUnmodifiableListView) return _achievements;
    // ignore: implicit_dynamic_type
    return EqualUnmodifiableListView(_achievements);
  }

  @override
  final String legend;
  final List<SeasonCardDto> _seasonHistory;
  @override
  @JsonKey(name: 'season_history')
  List<SeasonCardDto> get seasonHistory {
    if (_seasonHistory is EqualUnmodifiableListView) return _seasonHistory;
    // ignore: implicit_dynamic_type
    return EqualUnmodifiableListView(_seasonHistory);
  }

  @override
  String toString() {
    return 'MemberProfile(user: $user, stats: $stats, achievements: $achievements, legend: $legend, seasonHistory: $seasonHistory)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$MemberProfileImpl &&
            (identical(other.user, user) || other.user == user) &&
            (identical(other.stats, stats) || other.stats == stats) &&
            const DeepCollectionEquality()
                .equals(other._achievements, _achievements) &&
            (identical(other.legend, legend) || other.legend == legend) &&
            const DeepCollectionEquality()
                .equals(other._seasonHistory, _seasonHistory));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(
      runtimeType,
      user,
      stats,
      const DeepCollectionEquality().hash(_achievements),
      legend,
      const DeepCollectionEquality().hash(_seasonHistory));

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$MemberProfileImplCopyWith<_$MemberProfileImpl> get copyWith =>
      __$$MemberProfileImplCopyWithImpl<_$MemberProfileImpl>(this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$MemberProfileImplToJson(
      this,
    );
  }
}

abstract class _MemberProfile implements MemberProfile {
  const factory _MemberProfile(
      {required final ProfileUser user,
      required final UserStats stats,
      required final List<ProfileAchievement> achievements,
      required final String legend,
      @JsonKey(name: 'season_history')
      required final List<SeasonCardDto> seasonHistory}) = _$MemberProfileImpl;

  factory _MemberProfile.fromJson(Map<String, dynamic> json) =
      _$MemberProfileImpl.fromJson;

  @override
  ProfileUser get user;
  @override
  UserStats get stats;
  @override
  List<ProfileAchievement> get achievements;
  @override
  String get legend;
  @override
  @JsonKey(name: 'season_history')
  List<SeasonCardDto> get seasonHistory;
  @override
  @JsonKey(ignore: true)
  _$$MemberProfileImplCopyWith<_$MemberProfileImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

ProfileUser _$ProfileUserFromJson(Map<String, dynamic> json) {
  return _ProfileUser.fromJson(json);
}

/// @nodoc
mixin _$ProfileUser {
  String get username => throw _privateConstructorUsedError;
  @JsonKey(name: 'avatar_emoji')
  String? get avatarEmoji => throw _privateConstructorUsedError;
  @JsonKey(name: 'avatar_url')
  String? get avatarUrl => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $ProfileUserCopyWith<ProfileUser> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $ProfileUserCopyWith<$Res> {
  factory $ProfileUserCopyWith(
          ProfileUser value, $Res Function(ProfileUser) then) =
      _$ProfileUserCopyWithImpl<$Res, ProfileUser>;
  @useResult
  $Res call(
      {String username,
      @JsonKey(name: 'avatar_emoji') String? avatarEmoji,
      @JsonKey(name: 'avatar_url') String? avatarUrl});
}

/// @nodoc
class _$ProfileUserCopyWithImpl<$Res, $Val extends ProfileUser>
    implements $ProfileUserCopyWith<$Res> {
  _$ProfileUserCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? username = null,
    Object? avatarEmoji = freezed,
    Object? avatarUrl = freezed,
  }) {
    return _then(_value.copyWith(
      username: null == username
          ? _value.username
          : username // ignore: cast_nullable_to_non_nullable
              as String,
      avatarEmoji: freezed == avatarEmoji
          ? _value.avatarEmoji
          : avatarEmoji // ignore: cast_nullable_to_non_nullable
              as String?,
      avatarUrl: freezed == avatarUrl
          ? _value.avatarUrl
          : avatarUrl // ignore: cast_nullable_to_non_nullable
              as String?,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$ProfileUserImplCopyWith<$Res>
    implements $ProfileUserCopyWith<$Res> {
  factory _$$ProfileUserImplCopyWith(
          _$ProfileUserImpl value, $Res Function(_$ProfileUserImpl) then) =
      __$$ProfileUserImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {String username,
      @JsonKey(name: 'avatar_emoji') String? avatarEmoji,
      @JsonKey(name: 'avatar_url') String? avatarUrl});
}

/// @nodoc
class __$$ProfileUserImplCopyWithImpl<$Res>
    extends _$ProfileUserCopyWithImpl<$Res, _$ProfileUserImpl>
    implements _$$ProfileUserImplCopyWith<$Res> {
  __$$ProfileUserImplCopyWithImpl(
      _$ProfileUserImpl _value, $Res Function(_$ProfileUserImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? username = null,
    Object? avatarEmoji = freezed,
    Object? avatarUrl = freezed,
  }) {
    return _then(_$ProfileUserImpl(
      username: null == username
          ? _value.username
          : username // ignore: cast_nullable_to_non_nullable
              as String,
      avatarEmoji: freezed == avatarEmoji
          ? _value.avatarEmoji
          : avatarEmoji // ignore: cast_nullable_to_non_nullable
              as String?,
      avatarUrl: freezed == avatarUrl
          ? _value.avatarUrl
          : avatarUrl // ignore: cast_nullable_to_non_nullable
              as String?,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$ProfileUserImpl implements _ProfileUser {
  const _$ProfileUserImpl(
      {required this.username,
      @JsonKey(name: 'avatar_emoji') this.avatarEmoji,
      @JsonKey(name: 'avatar_url') this.avatarUrl});

  factory _$ProfileUserImpl.fromJson(Map<String, dynamic> json) =>
      _$$ProfileUserImplFromJson(json);

  @override
  final String username;
  @override
  @JsonKey(name: 'avatar_emoji')
  final String? avatarEmoji;
  @override
  @JsonKey(name: 'avatar_url')
  final String? avatarUrl;

  @override
  String toString() {
    return 'ProfileUser(username: $username, avatarEmoji: $avatarEmoji, avatarUrl: $avatarUrl)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$ProfileUserImpl &&
            (identical(other.username, username) ||
                other.username == username) &&
            (identical(other.avatarEmoji, avatarEmoji) ||
                other.avatarEmoji == avatarEmoji) &&
            (identical(other.avatarUrl, avatarUrl) ||
                other.avatarUrl == avatarUrl));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode =>
      Object.hash(runtimeType, username, avatarEmoji, avatarUrl);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$ProfileUserImplCopyWith<_$ProfileUserImpl> get copyWith =>
      __$$ProfileUserImplCopyWithImpl<_$ProfileUserImpl>(this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$ProfileUserImplToJson(
      this,
    );
  }
}

abstract class _ProfileUser implements ProfileUser {
  const factory _ProfileUser(
          {required final String username,
          @JsonKey(name: 'avatar_emoji') final String? avatarEmoji,
          @JsonKey(name: 'avatar_url') final String? avatarUrl}) =
      _$ProfileUserImpl;

  factory _ProfileUser.fromJson(Map<String, dynamic> json) =
      _$ProfileUserImpl.fromJson;

  @override
  String get username;
  @override
  @JsonKey(name: 'avatar_emoji')
  String? get avatarEmoji;
  @override
  @JsonKey(name: 'avatar_url')
  String? get avatarUrl;
  @override
  @JsonKey(ignore: true)
  _$$ProfileUserImplCopyWith<_$ProfileUserImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

UserStats _$UserStatsFromJson(Map<String, dynamic> json) {
  return _UserStats.fromJson(json);
}

/// @nodoc
mixin _$UserStats {
  @JsonKey(name: 'seasons_played')
  int get seasonsPlayed => throw _privateConstructorUsedError;
  @JsonKey(name: 'voting_streak')
  int get votingStreak => throw _privateConstructorUsedError;
  @JsonKey(name: 'max_voting_streak')
  int get maxVotingStreak => throw _privateConstructorUsedError;
  @JsonKey(name: 'guess_accuracy')
  double get guessAccuracy => throw _privateConstructorUsedError;
  @JsonKey(name: 'total_votes_cast')
  int get totalVotesCast => throw _privateConstructorUsedError;
  @JsonKey(name: 'total_votes_received')
  int get totalVotesReceived => throw _privateConstructorUsedError;
  @JsonKey(name: 'top_attribute_all_time')
  TopAttributeDto? get topAttributeAllTime =>
      throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $UserStatsCopyWith<UserStats> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $UserStatsCopyWith<$Res> {
  factory $UserStatsCopyWith(UserStats value, $Res Function(UserStats) then) =
      _$UserStatsCopyWithImpl<$Res, UserStats>;
  @useResult
  $Res call(
      {@JsonKey(name: 'seasons_played') int seasonsPlayed,
      @JsonKey(name: 'voting_streak') int votingStreak,
      @JsonKey(name: 'max_voting_streak') int maxVotingStreak,
      @JsonKey(name: 'guess_accuracy') double guessAccuracy,
      @JsonKey(name: 'total_votes_cast') int totalVotesCast,
      @JsonKey(name: 'total_votes_received') int totalVotesReceived,
      @JsonKey(name: 'top_attribute_all_time')
      TopAttributeDto? topAttributeAllTime});

  $TopAttributeDtoCopyWith<$Res>? get topAttributeAllTime;
}

/// @nodoc
class _$UserStatsCopyWithImpl<$Res, $Val extends UserStats>
    implements $UserStatsCopyWith<$Res> {
  _$UserStatsCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? seasonsPlayed = null,
    Object? votingStreak = null,
    Object? maxVotingStreak = null,
    Object? guessAccuracy = null,
    Object? totalVotesCast = null,
    Object? totalVotesReceived = null,
    Object? topAttributeAllTime = freezed,
  }) {
    return _then(_value.copyWith(
      seasonsPlayed: null == seasonsPlayed
          ? _value.seasonsPlayed
          : seasonsPlayed // ignore: cast_nullable_to_non_nullable
              as int,
      votingStreak: null == votingStreak
          ? _value.votingStreak
          : votingStreak // ignore: cast_nullable_to_non_nullable
              as int,
      maxVotingStreak: null == maxVotingStreak
          ? _value.maxVotingStreak
          : maxVotingStreak // ignore: cast_nullable_to_non_nullable
              as int,
      guessAccuracy: null == guessAccuracy
          ? _value.guessAccuracy
          : guessAccuracy // ignore: cast_nullable_to_non_nullable
              as double,
      totalVotesCast: null == totalVotesCast
          ? _value.totalVotesCast
          : totalVotesCast // ignore: cast_nullable_to_non_nullable
              as int,
      totalVotesReceived: null == totalVotesReceived
          ? _value.totalVotesReceived
          : totalVotesReceived // ignore: cast_nullable_to_non_nullable
              as int,
      topAttributeAllTime: freezed == topAttributeAllTime
          ? _value.topAttributeAllTime
          : topAttributeAllTime // ignore: cast_nullable_to_non_nullable
              as TopAttributeDto?,
    ) as $Val);
  }

  @override
  @pragma('vm:prefer-inline')
  $TopAttributeDtoCopyWith<$Res>? get topAttributeAllTime {
    if (_value.topAttributeAllTime == null) {
      return null;
    }

    return $TopAttributeDtoCopyWith<$Res>(_value.topAttributeAllTime!, (value) {
      return _then(_value.copyWith(topAttributeAllTime: value) as $Val);
    });
  }
}

/// @nodoc
abstract class _$$UserStatsImplCopyWith<$Res>
    implements $UserStatsCopyWith<$Res> {
  factory _$$UserStatsImplCopyWith(
          _$UserStatsImpl value, $Res Function(_$UserStatsImpl) then) =
      __$$UserStatsImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {@JsonKey(name: 'seasons_played') int seasonsPlayed,
      @JsonKey(name: 'voting_streak') int votingStreak,
      @JsonKey(name: 'max_voting_streak') int maxVotingStreak,
      @JsonKey(name: 'guess_accuracy') double guessAccuracy,
      @JsonKey(name: 'total_votes_cast') int totalVotesCast,
      @JsonKey(name: 'total_votes_received') int totalVotesReceived,
      @JsonKey(name: 'top_attribute_all_time')
      TopAttributeDto? topAttributeAllTime});

  @override
  $TopAttributeDtoCopyWith<$Res>? get topAttributeAllTime;
}

/// @nodoc
class __$$UserStatsImplCopyWithImpl<$Res>
    extends _$UserStatsCopyWithImpl<$Res, _$UserStatsImpl>
    implements _$$UserStatsImplCopyWith<$Res> {
  __$$UserStatsImplCopyWithImpl(
      _$UserStatsImpl _value, $Res Function(_$UserStatsImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? seasonsPlayed = null,
    Object? votingStreak = null,
    Object? maxVotingStreak = null,
    Object? guessAccuracy = null,
    Object? totalVotesCast = null,
    Object? totalVotesReceived = null,
    Object? topAttributeAllTime = freezed,
  }) {
    return _then(_$UserStatsImpl(
      seasonsPlayed: null == seasonsPlayed
          ? _value.seasonsPlayed
          : seasonsPlayed // ignore: cast_nullable_to_non_nullable
              as int,
      votingStreak: null == votingStreak
          ? _value.votingStreak
          : votingStreak // ignore: cast_nullable_to_non_nullable
              as int,
      maxVotingStreak: null == maxVotingStreak
          ? _value.maxVotingStreak
          : maxVotingStreak // ignore: cast_nullable_to_non_nullable
              as int,
      guessAccuracy: null == guessAccuracy
          ? _value.guessAccuracy
          : guessAccuracy // ignore: cast_nullable_to_non_nullable
              as double,
      totalVotesCast: null == totalVotesCast
          ? _value.totalVotesCast
          : totalVotesCast // ignore: cast_nullable_to_non_nullable
              as int,
      totalVotesReceived: null == totalVotesReceived
          ? _value.totalVotesReceived
          : totalVotesReceived // ignore: cast_nullable_to_non_nullable
              as int,
      topAttributeAllTime: freezed == topAttributeAllTime
          ? _value.topAttributeAllTime
          : topAttributeAllTime // ignore: cast_nullable_to_non_nullable
              as TopAttributeDto?,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$UserStatsImpl implements _UserStats {
  const _$UserStatsImpl(
      {@JsonKey(name: 'seasons_played') required this.seasonsPlayed,
      @JsonKey(name: 'voting_streak') required this.votingStreak,
      @JsonKey(name: 'max_voting_streak') required this.maxVotingStreak,
      @JsonKey(name: 'guess_accuracy') required this.guessAccuracy,
      @JsonKey(name: 'total_votes_cast') required this.totalVotesCast,
      @JsonKey(name: 'total_votes_received') required this.totalVotesReceived,
      @JsonKey(name: 'top_attribute_all_time') this.topAttributeAllTime});

  factory _$UserStatsImpl.fromJson(Map<String, dynamic> json) =>
      _$$UserStatsImplFromJson(json);

  @override
  @JsonKey(name: 'seasons_played')
  final int seasonsPlayed;
  @override
  @JsonKey(name: 'voting_streak')
  final int votingStreak;
  @override
  @JsonKey(name: 'max_voting_streak')
  final int maxVotingStreak;
  @override
  @JsonKey(name: 'guess_accuracy')
  final double guessAccuracy;
  @override
  @JsonKey(name: 'total_votes_cast')
  final int totalVotesCast;
  @override
  @JsonKey(name: 'total_votes_received')
  final int totalVotesReceived;
  @override
  @JsonKey(name: 'top_attribute_all_time')
  final TopAttributeDto? topAttributeAllTime;

  @override
  String toString() {
    return 'UserStats(seasonsPlayed: $seasonsPlayed, votingStreak: $votingStreak, maxVotingStreak: $maxVotingStreak, guessAccuracy: $guessAccuracy, totalVotesCast: $totalVotesCast, totalVotesReceived: $totalVotesReceived, topAttributeAllTime: $topAttributeAllTime)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$UserStatsImpl &&
            (identical(other.seasonsPlayed, seasonsPlayed) ||
                other.seasonsPlayed == seasonsPlayed) &&
            (identical(other.votingStreak, votingStreak) ||
                other.votingStreak == votingStreak) &&
            (identical(other.maxVotingStreak, maxVotingStreak) ||
                other.maxVotingStreak == maxVotingStreak) &&
            (identical(other.guessAccuracy, guessAccuracy) ||
                other.guessAccuracy == guessAccuracy) &&
            (identical(other.totalVotesCast, totalVotesCast) ||
                other.totalVotesCast == totalVotesCast) &&
            (identical(other.totalVotesReceived, totalVotesReceived) ||
                other.totalVotesReceived == totalVotesReceived) &&
            (identical(other.topAttributeAllTime, topAttributeAllTime) ||
                other.topAttributeAllTime == topAttributeAllTime));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(
      runtimeType,
      seasonsPlayed,
      votingStreak,
      maxVotingStreak,
      guessAccuracy,
      totalVotesCast,
      totalVotesReceived,
      topAttributeAllTime);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$UserStatsImplCopyWith<_$UserStatsImpl> get copyWith =>
      __$$UserStatsImplCopyWithImpl<_$UserStatsImpl>(this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$UserStatsImplToJson(
      this,
    );
  }
}

abstract class _UserStats implements UserStats {
  const factory _UserStats(
      {@JsonKey(name: 'seasons_played') required final int seasonsPlayed,
      @JsonKey(name: 'voting_streak') required final int votingStreak,
      @JsonKey(name: 'max_voting_streak') required final int maxVotingStreak,
      @JsonKey(name: 'guess_accuracy') required final double guessAccuracy,
      @JsonKey(name: 'total_votes_cast') required final int totalVotesCast,
      @JsonKey(name: 'total_votes_received')
      required final int totalVotesReceived,
      @JsonKey(name: 'top_attribute_all_time')
      final TopAttributeDto? topAttributeAllTime}) = _$UserStatsImpl;

  factory _UserStats.fromJson(Map<String, dynamic> json) =
      _$UserStatsImpl.fromJson;

  @override
  @JsonKey(name: 'seasons_played')
  int get seasonsPlayed;
  @override
  @JsonKey(name: 'voting_streak')
  int get votingStreak;
  @override
  @JsonKey(name: 'max_voting_streak')
  int get maxVotingStreak;
  @override
  @JsonKey(name: 'guess_accuracy')
  double get guessAccuracy;
  @override
  @JsonKey(name: 'total_votes_cast')
  int get totalVotesCast;
  @override
  @JsonKey(name: 'total_votes_received')
  int get totalVotesReceived;
  @override
  @JsonKey(name: 'top_attribute_all_time')
  TopAttributeDto? get topAttributeAllTime;
  @override
  @JsonKey(ignore: true)
  _$$UserStatsImplCopyWith<_$UserStatsImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

TopAttributeDto _$TopAttributeDtoFromJson(Map<String, dynamic> json) {
  return _TopAttributeDto.fromJson(json);
}

/// @nodoc
mixin _$TopAttributeDto {
  @JsonKey(name: 'question_text')
  String get questionText => throw _privateConstructorUsedError;
  double get percentage => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $TopAttributeDtoCopyWith<TopAttributeDto> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $TopAttributeDtoCopyWith<$Res> {
  factory $TopAttributeDtoCopyWith(
          TopAttributeDto value, $Res Function(TopAttributeDto) then) =
      _$TopAttributeDtoCopyWithImpl<$Res, TopAttributeDto>;
  @useResult
  $Res call(
      {@JsonKey(name: 'question_text') String questionText, double percentage});
}

/// @nodoc
class _$TopAttributeDtoCopyWithImpl<$Res, $Val extends TopAttributeDto>
    implements $TopAttributeDtoCopyWith<$Res> {
  _$TopAttributeDtoCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? questionText = null,
    Object? percentage = null,
  }) {
    return _then(_value.copyWith(
      questionText: null == questionText
          ? _value.questionText
          : questionText // ignore: cast_nullable_to_non_nullable
              as String,
      percentage: null == percentage
          ? _value.percentage
          : percentage // ignore: cast_nullable_to_non_nullable
              as double,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$TopAttributeDtoImplCopyWith<$Res>
    implements $TopAttributeDtoCopyWith<$Res> {
  factory _$$TopAttributeDtoImplCopyWith(_$TopAttributeDtoImpl value,
          $Res Function(_$TopAttributeDtoImpl) then) =
      __$$TopAttributeDtoImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {@JsonKey(name: 'question_text') String questionText, double percentage});
}

/// @nodoc
class __$$TopAttributeDtoImplCopyWithImpl<$Res>
    extends _$TopAttributeDtoCopyWithImpl<$Res, _$TopAttributeDtoImpl>
    implements _$$TopAttributeDtoImplCopyWith<$Res> {
  __$$TopAttributeDtoImplCopyWithImpl(
      _$TopAttributeDtoImpl _value, $Res Function(_$TopAttributeDtoImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? questionText = null,
    Object? percentage = null,
  }) {
    return _then(_$TopAttributeDtoImpl(
      questionText: null == questionText
          ? _value.questionText
          : questionText // ignore: cast_nullable_to_non_nullable
              as String,
      percentage: null == percentage
          ? _value.percentage
          : percentage // ignore: cast_nullable_to_non_nullable
              as double,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$TopAttributeDtoImpl implements _TopAttributeDto {
  const _$TopAttributeDtoImpl(
      {@JsonKey(name: 'question_text') required this.questionText,
      required this.percentage});

  factory _$TopAttributeDtoImpl.fromJson(Map<String, dynamic> json) =>
      _$$TopAttributeDtoImplFromJson(json);

  @override
  @JsonKey(name: 'question_text')
  final String questionText;
  @override
  final double percentage;

  @override
  String toString() {
    return 'TopAttributeDto(questionText: $questionText, percentage: $percentage)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$TopAttributeDtoImpl &&
            (identical(other.questionText, questionText) ||
                other.questionText == questionText) &&
            (identical(other.percentage, percentage) ||
                other.percentage == percentage));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(runtimeType, questionText, percentage);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$TopAttributeDtoImplCopyWith<_$TopAttributeDtoImpl> get copyWith =>
      __$$TopAttributeDtoImplCopyWithImpl<_$TopAttributeDtoImpl>(
          this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$TopAttributeDtoImplToJson(
      this,
    );
  }
}

abstract class _TopAttributeDto implements TopAttributeDto {
  const factory _TopAttributeDto(
      {@JsonKey(name: 'question_text') required final String questionText,
      required final double percentage}) = _$TopAttributeDtoImpl;

  factory _TopAttributeDto.fromJson(Map<String, dynamic> json) =
      _$TopAttributeDtoImpl.fromJson;

  @override
  @JsonKey(name: 'question_text')
  String get questionText;
  @override
  double get percentage;
  @override
  @JsonKey(ignore: true)
  _$$TopAttributeDtoImplCopyWith<_$TopAttributeDtoImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

ProfileAchievement _$ProfileAchievementFromJson(Map<String, dynamic> json) {
  return _ProfileAchievement.fromJson(json);
}

/// @nodoc
mixin _$ProfileAchievement {
  String get type => throw _privateConstructorUsedError;
  Map<String, dynamic>? get metadata => throw _privateConstructorUsedError;
  @JsonKey(name: 'earned_at')
  String get earnedAt => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $ProfileAchievementCopyWith<ProfileAchievement> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $ProfileAchievementCopyWith<$Res> {
  factory $ProfileAchievementCopyWith(
          ProfileAchievement value, $Res Function(ProfileAchievement) then) =
      _$ProfileAchievementCopyWithImpl<$Res, ProfileAchievement>;
  @useResult
  $Res call(
      {String type,
      Map<String, dynamic>? metadata,
      @JsonKey(name: 'earned_at') String earnedAt});
}

/// @nodoc
class _$ProfileAchievementCopyWithImpl<$Res, $Val extends ProfileAchievement>
    implements $ProfileAchievementCopyWith<$Res> {
  _$ProfileAchievementCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? type = null,
    Object? metadata = freezed,
    Object? earnedAt = null,
  }) {
    return _then(_value.copyWith(
      type: null == type
          ? _value.type
          : type // ignore: cast_nullable_to_non_nullable
              as String,
      metadata: freezed == metadata
          ? _value.metadata
          : metadata // ignore: cast_nullable_to_non_nullable
              as Map<String, dynamic>?,
      earnedAt: null == earnedAt
          ? _value.earnedAt
          : earnedAt // ignore: cast_nullable_to_non_nullable
              as String,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$ProfileAchievementImplCopyWith<$Res>
    implements $ProfileAchievementCopyWith<$Res> {
  factory _$$ProfileAchievementImplCopyWith(_$ProfileAchievementImpl value,
          $Res Function(_$ProfileAchievementImpl) then) =
      __$$ProfileAchievementImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {String type,
      Map<String, dynamic>? metadata,
      @JsonKey(name: 'earned_at') String earnedAt});
}

/// @nodoc
class __$$ProfileAchievementImplCopyWithImpl<$Res>
    extends _$ProfileAchievementCopyWithImpl<$Res, _$ProfileAchievementImpl>
    implements _$$ProfileAchievementImplCopyWith<$Res> {
  __$$ProfileAchievementImplCopyWithImpl(_$ProfileAchievementImpl _value,
      $Res Function(_$ProfileAchievementImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? type = null,
    Object? metadata = freezed,
    Object? earnedAt = null,
  }) {
    return _then(_$ProfileAchievementImpl(
      type: null == type
          ? _value.type
          : type // ignore: cast_nullable_to_non_nullable
              as String,
      metadata: freezed == metadata
          ? _value._metadata
          : metadata // ignore: cast_nullable_to_non_nullable
              as Map<String, dynamic>?,
      earnedAt: null == earnedAt
          ? _value.earnedAt
          : earnedAt // ignore: cast_nullable_to_non_nullable
              as String,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$ProfileAchievementImpl implements _ProfileAchievement {
  const _$ProfileAchievementImpl(
      {required this.type,
      final Map<String, dynamic>? metadata,
      @JsonKey(name: 'earned_at') required this.earnedAt})
      : _metadata = metadata;

  factory _$ProfileAchievementImpl.fromJson(Map<String, dynamic> json) =>
      _$$ProfileAchievementImplFromJson(json);

  @override
  final String type;
  final Map<String, dynamic>? _metadata;
  @override
  Map<String, dynamic>? get metadata {
    final value = _metadata;
    if (value == null) return null;
    if (_metadata is EqualUnmodifiableMapView) return _metadata;
    // ignore: implicit_dynamic_type
    return EqualUnmodifiableMapView(value);
  }

  @override
  @JsonKey(name: 'earned_at')
  final String earnedAt;

  @override
  String toString() {
    return 'ProfileAchievement(type: $type, metadata: $metadata, earnedAt: $earnedAt)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$ProfileAchievementImpl &&
            (identical(other.type, type) || other.type == type) &&
            const DeepCollectionEquality().equals(other._metadata, _metadata) &&
            (identical(other.earnedAt, earnedAt) ||
                other.earnedAt == earnedAt));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(runtimeType, type,
      const DeepCollectionEquality().hash(_metadata), earnedAt);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$ProfileAchievementImplCopyWith<_$ProfileAchievementImpl> get copyWith =>
      __$$ProfileAchievementImplCopyWithImpl<_$ProfileAchievementImpl>(
          this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$ProfileAchievementImplToJson(
      this,
    );
  }
}

abstract class _ProfileAchievement implements ProfileAchievement {
  const factory _ProfileAchievement(
          {required final String type,
          final Map<String, dynamic>? metadata,
          @JsonKey(name: 'earned_at') required final String earnedAt}) =
      _$ProfileAchievementImpl;

  factory _ProfileAchievement.fromJson(Map<String, dynamic> json) =
      _$ProfileAchievementImpl.fromJson;

  @override
  String get type;
  @override
  Map<String, dynamic>? get metadata;
  @override
  @JsonKey(name: 'earned_at')
  String get earnedAt;
  @override
  @JsonKey(ignore: true)
  _$$ProfileAchievementImplCopyWith<_$ProfileAchievementImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

SeasonCardDto _$SeasonCardDtoFromJson(Map<String, dynamic> json) {
  return _SeasonCardDto.fromJson(json);
}

/// @nodoc
mixin _$SeasonCardDto {
  @JsonKey(name: 'season_id')
  String get seasonId => throw _privateConstructorUsedError;
  @JsonKey(name: 'season_number')
  int get seasonNumber => throw _privateConstructorUsedError;
  @JsonKey(name: 'top_attribute')
  String get topAttribute => throw _privateConstructorUsedError;
  String get category => throw _privateConstructorUsedError;
  double get percentage => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $SeasonCardDtoCopyWith<SeasonCardDto> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $SeasonCardDtoCopyWith<$Res> {
  factory $SeasonCardDtoCopyWith(
          SeasonCardDto value, $Res Function(SeasonCardDto) then) =
      _$SeasonCardDtoCopyWithImpl<$Res, SeasonCardDto>;
  @useResult
  $Res call(
      {@JsonKey(name: 'season_id') String seasonId,
      @JsonKey(name: 'season_number') int seasonNumber,
      @JsonKey(name: 'top_attribute') String topAttribute,
      String category,
      double percentage});
}

/// @nodoc
class _$SeasonCardDtoCopyWithImpl<$Res, $Val extends SeasonCardDto>
    implements $SeasonCardDtoCopyWith<$Res> {
  _$SeasonCardDtoCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? seasonId = null,
    Object? seasonNumber = null,
    Object? topAttribute = null,
    Object? category = null,
    Object? percentage = null,
  }) {
    return _then(_value.copyWith(
      seasonId: null == seasonId
          ? _value.seasonId
          : seasonId // ignore: cast_nullable_to_non_nullable
              as String,
      seasonNumber: null == seasonNumber
          ? _value.seasonNumber
          : seasonNumber // ignore: cast_nullable_to_non_nullable
              as int,
      topAttribute: null == topAttribute
          ? _value.topAttribute
          : topAttribute // ignore: cast_nullable_to_non_nullable
              as String,
      category: null == category
          ? _value.category
          : category // ignore: cast_nullable_to_non_nullable
              as String,
      percentage: null == percentage
          ? _value.percentage
          : percentage // ignore: cast_nullable_to_non_nullable
              as double,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$SeasonCardDtoImplCopyWith<$Res>
    implements $SeasonCardDtoCopyWith<$Res> {
  factory _$$SeasonCardDtoImplCopyWith(
          _$SeasonCardDtoImpl value, $Res Function(_$SeasonCardDtoImpl) then) =
      __$$SeasonCardDtoImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {@JsonKey(name: 'season_id') String seasonId,
      @JsonKey(name: 'season_number') int seasonNumber,
      @JsonKey(name: 'top_attribute') String topAttribute,
      String category,
      double percentage});
}

/// @nodoc
class __$$SeasonCardDtoImplCopyWithImpl<$Res>
    extends _$SeasonCardDtoCopyWithImpl<$Res, _$SeasonCardDtoImpl>
    implements _$$SeasonCardDtoImplCopyWith<$Res> {
  __$$SeasonCardDtoImplCopyWithImpl(
      _$SeasonCardDtoImpl _value, $Res Function(_$SeasonCardDtoImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? seasonId = null,
    Object? seasonNumber = null,
    Object? topAttribute = null,
    Object? category = null,
    Object? percentage = null,
  }) {
    return _then(_$SeasonCardDtoImpl(
      seasonId: null == seasonId
          ? _value.seasonId
          : seasonId // ignore: cast_nullable_to_non_nullable
              as String,
      seasonNumber: null == seasonNumber
          ? _value.seasonNumber
          : seasonNumber // ignore: cast_nullable_to_non_nullable
              as int,
      topAttribute: null == topAttribute
          ? _value.topAttribute
          : topAttribute // ignore: cast_nullable_to_non_nullable
              as String,
      category: null == category
          ? _value.category
          : category // ignore: cast_nullable_to_non_nullable
              as String,
      percentage: null == percentage
          ? _value.percentage
          : percentage // ignore: cast_nullable_to_non_nullable
              as double,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$SeasonCardDtoImpl implements _SeasonCardDto {
  const _$SeasonCardDtoImpl(
      {@JsonKey(name: 'season_id') required this.seasonId,
      @JsonKey(name: 'season_number') required this.seasonNumber,
      @JsonKey(name: 'top_attribute') required this.topAttribute,
      required this.category,
      required this.percentage});

  factory _$SeasonCardDtoImpl.fromJson(Map<String, dynamic> json) =>
      _$$SeasonCardDtoImplFromJson(json);

  @override
  @JsonKey(name: 'season_id')
  final String seasonId;
  @override
  @JsonKey(name: 'season_number')
  final int seasonNumber;
  @override
  @JsonKey(name: 'top_attribute')
  final String topAttribute;
  @override
  final String category;
  @override
  final double percentage;

  @override
  String toString() {
    return 'SeasonCardDto(seasonId: $seasonId, seasonNumber: $seasonNumber, topAttribute: $topAttribute, category: $category, percentage: $percentage)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$SeasonCardDtoImpl &&
            (identical(other.seasonId, seasonId) ||
                other.seasonId == seasonId) &&
            (identical(other.seasonNumber, seasonNumber) ||
                other.seasonNumber == seasonNumber) &&
            (identical(other.topAttribute, topAttribute) ||
                other.topAttribute == topAttribute) &&
            (identical(other.category, category) ||
                other.category == category) &&
            (identical(other.percentage, percentage) ||
                other.percentage == percentage));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(
      runtimeType, seasonId, seasonNumber, topAttribute, category, percentage);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$SeasonCardDtoImplCopyWith<_$SeasonCardDtoImpl> get copyWith =>
      __$$SeasonCardDtoImplCopyWithImpl<_$SeasonCardDtoImpl>(this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$SeasonCardDtoImplToJson(
      this,
    );
  }
}

abstract class _SeasonCardDto implements SeasonCardDto {
  const factory _SeasonCardDto(
      {@JsonKey(name: 'season_id') required final String seasonId,
      @JsonKey(name: 'season_number') required final int seasonNumber,
      @JsonKey(name: 'top_attribute') required final String topAttribute,
      required final String category,
      required final double percentage}) = _$SeasonCardDtoImpl;

  factory _SeasonCardDto.fromJson(Map<String, dynamic> json) =
      _$SeasonCardDtoImpl.fromJson;

  @override
  @JsonKey(name: 'season_id')
  String get seasonId;
  @override
  @JsonKey(name: 'season_number')
  int get seasonNumber;
  @override
  @JsonKey(name: 'top_attribute')
  String get topAttribute;
  @override
  String get category;
  @override
  double get percentage;
  @override
  @JsonKey(ignore: true)
  _$$SeasonCardDtoImplCopyWith<_$SeasonCardDtoImpl> get copyWith =>
      throw _privateConstructorUsedError;
}
