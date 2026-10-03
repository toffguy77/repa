// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'chronicle.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

T _$identity<T>(T value) => value;

final _privateConstructorUsedError = UnsupportedError(
    'It seems like you constructed your class using `MyClass._()`. This constructor is only meant to be used by freezed and you are not supposed to need it nor use it.\nPlease check the documentation here for more information: https://github.com/rrousselGit/freezed#adding-getters-and-methods-to-our-models');

ChronicleEntry _$ChronicleEntryFromJson(Map<String, dynamic> json) {
  return _ChronicleEntry.fromJson(json);
}

/// @nodoc
mixin _$ChronicleEntry {
  @JsonKey(name: 'question_id')
  String get questionId => throw _privateConstructorUsedError;
  @JsonKey(name: 'question_text')
  String get questionText => throw _privateConstructorUsedError;
  String get category => throw _privateConstructorUsedError;
  @JsonKey(name: 'user_id')
  String get userId => throw _privateConstructorUsedError;
  String get username => throw _privateConstructorUsedError;
  @JsonKey(name: 'avatar_emoji')
  String? get avatarEmoji => throw _privateConstructorUsedError;
  double get percentage => throw _privateConstructorUsedError;
  @JsonKey(name: 'vote_count')
  int get voteCount => throw _privateConstructorUsedError;
  @JsonKey(name: 'total_voters')
  int get totalVoters => throw _privateConstructorUsedError;

  /// Whether this result's percentages were computed over too few voters to hide who voted. The
  /// server decides, using the same threshold the detector is gated on — the client must not carry
  /// its own copy of that number.
  @JsonKey(name: 'small_sample')
  bool get smallSample => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $ChronicleEntryCopyWith<ChronicleEntry> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $ChronicleEntryCopyWith<$Res> {
  factory $ChronicleEntryCopyWith(
          ChronicleEntry value, $Res Function(ChronicleEntry) then) =
      _$ChronicleEntryCopyWithImpl<$Res, ChronicleEntry>;
  @useResult
  $Res call(
      {@JsonKey(name: 'question_id') String questionId,
      @JsonKey(name: 'question_text') String questionText,
      String category,
      @JsonKey(name: 'user_id') String userId,
      String username,
      @JsonKey(name: 'avatar_emoji') String? avatarEmoji,
      double percentage,
      @JsonKey(name: 'vote_count') int voteCount,
      @JsonKey(name: 'total_voters') int totalVoters,
      @JsonKey(name: 'small_sample') bool smallSample});
}

/// @nodoc
class _$ChronicleEntryCopyWithImpl<$Res, $Val extends ChronicleEntry>
    implements $ChronicleEntryCopyWith<$Res> {
  _$ChronicleEntryCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? questionId = null,
    Object? questionText = null,
    Object? category = null,
    Object? userId = null,
    Object? username = null,
    Object? avatarEmoji = freezed,
    Object? percentage = null,
    Object? voteCount = null,
    Object? totalVoters = null,
    Object? smallSample = null,
  }) {
    return _then(_value.copyWith(
      questionId: null == questionId
          ? _value.questionId
          : questionId // ignore: cast_nullable_to_non_nullable
              as String,
      questionText: null == questionText
          ? _value.questionText
          : questionText // ignore: cast_nullable_to_non_nullable
              as String,
      category: null == category
          ? _value.category
          : category // ignore: cast_nullable_to_non_nullable
              as String,
      userId: null == userId
          ? _value.userId
          : userId // ignore: cast_nullable_to_non_nullable
              as String,
      username: null == username
          ? _value.username
          : username // ignore: cast_nullable_to_non_nullable
              as String,
      avatarEmoji: freezed == avatarEmoji
          ? _value.avatarEmoji
          : avatarEmoji // ignore: cast_nullable_to_non_nullable
              as String?,
      percentage: null == percentage
          ? _value.percentage
          : percentage // ignore: cast_nullable_to_non_nullable
              as double,
      voteCount: null == voteCount
          ? _value.voteCount
          : voteCount // ignore: cast_nullable_to_non_nullable
              as int,
      totalVoters: null == totalVoters
          ? _value.totalVoters
          : totalVoters // ignore: cast_nullable_to_non_nullable
              as int,
      smallSample: null == smallSample
          ? _value.smallSample
          : smallSample // ignore: cast_nullable_to_non_nullable
              as bool,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$ChronicleEntryImplCopyWith<$Res>
    implements $ChronicleEntryCopyWith<$Res> {
  factory _$$ChronicleEntryImplCopyWith(_$ChronicleEntryImpl value,
          $Res Function(_$ChronicleEntryImpl) then) =
      __$$ChronicleEntryImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {@JsonKey(name: 'question_id') String questionId,
      @JsonKey(name: 'question_text') String questionText,
      String category,
      @JsonKey(name: 'user_id') String userId,
      String username,
      @JsonKey(name: 'avatar_emoji') String? avatarEmoji,
      double percentage,
      @JsonKey(name: 'vote_count') int voteCount,
      @JsonKey(name: 'total_voters') int totalVoters,
      @JsonKey(name: 'small_sample') bool smallSample});
}

/// @nodoc
class __$$ChronicleEntryImplCopyWithImpl<$Res>
    extends _$ChronicleEntryCopyWithImpl<$Res, _$ChronicleEntryImpl>
    implements _$$ChronicleEntryImplCopyWith<$Res> {
  __$$ChronicleEntryImplCopyWithImpl(
      _$ChronicleEntryImpl _value, $Res Function(_$ChronicleEntryImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? questionId = null,
    Object? questionText = null,
    Object? category = null,
    Object? userId = null,
    Object? username = null,
    Object? avatarEmoji = freezed,
    Object? percentage = null,
    Object? voteCount = null,
    Object? totalVoters = null,
    Object? smallSample = null,
  }) {
    return _then(_$ChronicleEntryImpl(
      questionId: null == questionId
          ? _value.questionId
          : questionId // ignore: cast_nullable_to_non_nullable
              as String,
      questionText: null == questionText
          ? _value.questionText
          : questionText // ignore: cast_nullable_to_non_nullable
              as String,
      category: null == category
          ? _value.category
          : category // ignore: cast_nullable_to_non_nullable
              as String,
      userId: null == userId
          ? _value.userId
          : userId // ignore: cast_nullable_to_non_nullable
              as String,
      username: null == username
          ? _value.username
          : username // ignore: cast_nullable_to_non_nullable
              as String,
      avatarEmoji: freezed == avatarEmoji
          ? _value.avatarEmoji
          : avatarEmoji // ignore: cast_nullable_to_non_nullable
              as String?,
      percentage: null == percentage
          ? _value.percentage
          : percentage // ignore: cast_nullable_to_non_nullable
              as double,
      voteCount: null == voteCount
          ? _value.voteCount
          : voteCount // ignore: cast_nullable_to_non_nullable
              as int,
      totalVoters: null == totalVoters
          ? _value.totalVoters
          : totalVoters // ignore: cast_nullable_to_non_nullable
              as int,
      smallSample: null == smallSample
          ? _value.smallSample
          : smallSample // ignore: cast_nullable_to_non_nullable
              as bool,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$ChronicleEntryImpl implements _ChronicleEntry {
  const _$ChronicleEntryImpl(
      {@JsonKey(name: 'question_id') required this.questionId,
      @JsonKey(name: 'question_text') required this.questionText,
      required this.category,
      @JsonKey(name: 'user_id') required this.userId,
      required this.username,
      @JsonKey(name: 'avatar_emoji') this.avatarEmoji,
      required this.percentage,
      @JsonKey(name: 'vote_count') required this.voteCount,
      @JsonKey(name: 'total_voters') required this.totalVoters,
      @JsonKey(name: 'small_sample') this.smallSample = false});

  factory _$ChronicleEntryImpl.fromJson(Map<String, dynamic> json) =>
      _$$ChronicleEntryImplFromJson(json);

  @override
  @JsonKey(name: 'question_id')
  final String questionId;
  @override
  @JsonKey(name: 'question_text')
  final String questionText;
  @override
  final String category;
  @override
  @JsonKey(name: 'user_id')
  final String userId;
  @override
  final String username;
  @override
  @JsonKey(name: 'avatar_emoji')
  final String? avatarEmoji;
  @override
  final double percentage;
  @override
  @JsonKey(name: 'vote_count')
  final int voteCount;
  @override
  @JsonKey(name: 'total_voters')
  final int totalVoters;

  /// Whether this result's percentages were computed over too few voters to hide who voted. The
  /// server decides, using the same threshold the detector is gated on — the client must not carry
  /// its own copy of that number.
  @override
  @JsonKey(name: 'small_sample')
  final bool smallSample;

  @override
  String toString() {
    return 'ChronicleEntry(questionId: $questionId, questionText: $questionText, category: $category, userId: $userId, username: $username, avatarEmoji: $avatarEmoji, percentage: $percentage, voteCount: $voteCount, totalVoters: $totalVoters, smallSample: $smallSample)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$ChronicleEntryImpl &&
            (identical(other.questionId, questionId) ||
                other.questionId == questionId) &&
            (identical(other.questionText, questionText) ||
                other.questionText == questionText) &&
            (identical(other.category, category) ||
                other.category == category) &&
            (identical(other.userId, userId) || other.userId == userId) &&
            (identical(other.username, username) ||
                other.username == username) &&
            (identical(other.avatarEmoji, avatarEmoji) ||
                other.avatarEmoji == avatarEmoji) &&
            (identical(other.percentage, percentage) ||
                other.percentage == percentage) &&
            (identical(other.voteCount, voteCount) ||
                other.voteCount == voteCount) &&
            (identical(other.totalVoters, totalVoters) ||
                other.totalVoters == totalVoters) &&
            (identical(other.smallSample, smallSample) ||
                other.smallSample == smallSample));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(
      runtimeType,
      questionId,
      questionText,
      category,
      userId,
      username,
      avatarEmoji,
      percentage,
      voteCount,
      totalVoters,
      smallSample);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$ChronicleEntryImplCopyWith<_$ChronicleEntryImpl> get copyWith =>
      __$$ChronicleEntryImplCopyWithImpl<_$ChronicleEntryImpl>(
          this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$ChronicleEntryImplToJson(
      this,
    );
  }
}

abstract class _ChronicleEntry implements ChronicleEntry {
  const factory _ChronicleEntry(
          {@JsonKey(name: 'question_id') required final String questionId,
          @JsonKey(name: 'question_text') required final String questionText,
          required final String category,
          @JsonKey(name: 'user_id') required final String userId,
          required final String username,
          @JsonKey(name: 'avatar_emoji') final String? avatarEmoji,
          required final double percentage,
          @JsonKey(name: 'vote_count') required final int voteCount,
          @JsonKey(name: 'total_voters') required final int totalVoters,
          @JsonKey(name: 'small_sample') final bool smallSample}) =
      _$ChronicleEntryImpl;

  factory _ChronicleEntry.fromJson(Map<String, dynamic> json) =
      _$ChronicleEntryImpl.fromJson;

  @override
  @JsonKey(name: 'question_id')
  String get questionId;
  @override
  @JsonKey(name: 'question_text')
  String get questionText;
  @override
  String get category;
  @override
  @JsonKey(name: 'user_id')
  String get userId;
  @override
  String get username;
  @override
  @JsonKey(name: 'avatar_emoji')
  String? get avatarEmoji;
  @override
  double get percentage;
  @override
  @JsonKey(name: 'vote_count')
  int get voteCount;
  @override
  @JsonKey(name: 'total_voters')
  int get totalVoters;
  @override

  /// Whether this result's percentages were computed over too few voters to hide who voted. The
  /// server decides, using the same threshold the detector is gated on — the client must not carry
  /// its own copy of that number.
  @JsonKey(name: 'small_sample')
  bool get smallSample;
  @override
  @JsonKey(ignore: true)
  _$$ChronicleEntryImplCopyWith<_$ChronicleEntryImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

ChronicleSeason _$ChronicleSeasonFromJson(Map<String, dynamic> json) {
  return _ChronicleSeason.fromJson(json);
}

/// @nodoc
mixin _$ChronicleSeason {
  @JsonKey(name: 'season_id')
  String get seasonId => throw _privateConstructorUsedError;
  int get number => throw _privateConstructorUsedError;
  @JsonKey(name: 'reveal_at')
  String get revealAt => throw _privateConstructorUsedError;
  List<ChronicleEntry> get entries => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $ChronicleSeasonCopyWith<ChronicleSeason> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $ChronicleSeasonCopyWith<$Res> {
  factory $ChronicleSeasonCopyWith(
          ChronicleSeason value, $Res Function(ChronicleSeason) then) =
      _$ChronicleSeasonCopyWithImpl<$Res, ChronicleSeason>;
  @useResult
  $Res call(
      {@JsonKey(name: 'season_id') String seasonId,
      int number,
      @JsonKey(name: 'reveal_at') String revealAt,
      List<ChronicleEntry> entries});
}

/// @nodoc
class _$ChronicleSeasonCopyWithImpl<$Res, $Val extends ChronicleSeason>
    implements $ChronicleSeasonCopyWith<$Res> {
  _$ChronicleSeasonCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? seasonId = null,
    Object? number = null,
    Object? revealAt = null,
    Object? entries = null,
  }) {
    return _then(_value.copyWith(
      seasonId: null == seasonId
          ? _value.seasonId
          : seasonId // ignore: cast_nullable_to_non_nullable
              as String,
      number: null == number
          ? _value.number
          : number // ignore: cast_nullable_to_non_nullable
              as int,
      revealAt: null == revealAt
          ? _value.revealAt
          : revealAt // ignore: cast_nullable_to_non_nullable
              as String,
      entries: null == entries
          ? _value.entries
          : entries // ignore: cast_nullable_to_non_nullable
              as List<ChronicleEntry>,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$ChronicleSeasonImplCopyWith<$Res>
    implements $ChronicleSeasonCopyWith<$Res> {
  factory _$$ChronicleSeasonImplCopyWith(_$ChronicleSeasonImpl value,
          $Res Function(_$ChronicleSeasonImpl) then) =
      __$$ChronicleSeasonImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {@JsonKey(name: 'season_id') String seasonId,
      int number,
      @JsonKey(name: 'reveal_at') String revealAt,
      List<ChronicleEntry> entries});
}

/// @nodoc
class __$$ChronicleSeasonImplCopyWithImpl<$Res>
    extends _$ChronicleSeasonCopyWithImpl<$Res, _$ChronicleSeasonImpl>
    implements _$$ChronicleSeasonImplCopyWith<$Res> {
  __$$ChronicleSeasonImplCopyWithImpl(
      _$ChronicleSeasonImpl _value, $Res Function(_$ChronicleSeasonImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? seasonId = null,
    Object? number = null,
    Object? revealAt = null,
    Object? entries = null,
  }) {
    return _then(_$ChronicleSeasonImpl(
      seasonId: null == seasonId
          ? _value.seasonId
          : seasonId // ignore: cast_nullable_to_non_nullable
              as String,
      number: null == number
          ? _value.number
          : number // ignore: cast_nullable_to_non_nullable
              as int,
      revealAt: null == revealAt
          ? _value.revealAt
          : revealAt // ignore: cast_nullable_to_non_nullable
              as String,
      entries: null == entries
          ? _value._entries
          : entries // ignore: cast_nullable_to_non_nullable
              as List<ChronicleEntry>,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$ChronicleSeasonImpl implements _ChronicleSeason {
  const _$ChronicleSeasonImpl(
      {@JsonKey(name: 'season_id') required this.seasonId,
      required this.number,
      @JsonKey(name: 'reveal_at') required this.revealAt,
      final List<ChronicleEntry> entries = const []})
      : _entries = entries;

  factory _$ChronicleSeasonImpl.fromJson(Map<String, dynamic> json) =>
      _$$ChronicleSeasonImplFromJson(json);

  @override
  @JsonKey(name: 'season_id')
  final String seasonId;
  @override
  final int number;
  @override
  @JsonKey(name: 'reveal_at')
  final String revealAt;
  final List<ChronicleEntry> _entries;
  @override
  @JsonKey()
  List<ChronicleEntry> get entries {
    if (_entries is EqualUnmodifiableListView) return _entries;
    // ignore: implicit_dynamic_type
    return EqualUnmodifiableListView(_entries);
  }

  @override
  String toString() {
    return 'ChronicleSeason(seasonId: $seasonId, number: $number, revealAt: $revealAt, entries: $entries)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$ChronicleSeasonImpl &&
            (identical(other.seasonId, seasonId) ||
                other.seasonId == seasonId) &&
            (identical(other.number, number) || other.number == number) &&
            (identical(other.revealAt, revealAt) ||
                other.revealAt == revealAt) &&
            const DeepCollectionEquality().equals(other._entries, _entries));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(runtimeType, seasonId, number, revealAt,
      const DeepCollectionEquality().hash(_entries));

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$ChronicleSeasonImplCopyWith<_$ChronicleSeasonImpl> get copyWith =>
      __$$ChronicleSeasonImplCopyWithImpl<_$ChronicleSeasonImpl>(
          this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$ChronicleSeasonImplToJson(
      this,
    );
  }
}

abstract class _ChronicleSeason implements ChronicleSeason {
  const factory _ChronicleSeason(
      {@JsonKey(name: 'season_id') required final String seasonId,
      required final int number,
      @JsonKey(name: 'reveal_at') required final String revealAt,
      final List<ChronicleEntry> entries}) = _$ChronicleSeasonImpl;

  factory _ChronicleSeason.fromJson(Map<String, dynamic> json) =
      _$ChronicleSeasonImpl.fromJson;

  @override
  @JsonKey(name: 'season_id')
  String get seasonId;
  @override
  int get number;
  @override
  @JsonKey(name: 'reveal_at')
  String get revealAt;
  @override
  List<ChronicleEntry> get entries;
  @override
  @JsonKey(ignore: true)
  _$$ChronicleSeasonImplCopyWith<_$ChronicleSeasonImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

StandingEntry _$StandingEntryFromJson(Map<String, dynamic> json) {
  return _StandingEntry.fromJson(json);
}

/// @nodoc
mixin _$StandingEntry {
  @JsonKey(name: 'user_id')
  String get userId => throw _privateConstructorUsedError;
  String get username => throw _privateConstructorUsedError;
  @JsonKey(name: 'avatar_emoji')
  String? get avatarEmoji => throw _privateConstructorUsedError;

  /// 1-based, or 0 for a member who has not been through a reveal yet.
  int get rank => throw _privateConstructorUsedError;

  /// Whether this member has a place at all. "Has not played here yet" and "placed last" are
  /// different facts, and showing the first as the second says something untrue about a newcomer.
  bool get ranked => throw _privateConstructorUsedError;

  /// How much history the place is based on: a place over two seasons and one over ten are not
  /// comparable, so the basis is shown next to the position.
  @JsonKey(name: 'seasons_played')
  int get seasonsPlayed => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $StandingEntryCopyWith<StandingEntry> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $StandingEntryCopyWith<$Res> {
  factory $StandingEntryCopyWith(
          StandingEntry value, $Res Function(StandingEntry) then) =
      _$StandingEntryCopyWithImpl<$Res, StandingEntry>;
  @useResult
  $Res call(
      {@JsonKey(name: 'user_id') String userId,
      String username,
      @JsonKey(name: 'avatar_emoji') String? avatarEmoji,
      int rank,
      bool ranked,
      @JsonKey(name: 'seasons_played') int seasonsPlayed});
}

/// @nodoc
class _$StandingEntryCopyWithImpl<$Res, $Val extends StandingEntry>
    implements $StandingEntryCopyWith<$Res> {
  _$StandingEntryCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? userId = null,
    Object? username = null,
    Object? avatarEmoji = freezed,
    Object? rank = null,
    Object? ranked = null,
    Object? seasonsPlayed = null,
  }) {
    return _then(_value.copyWith(
      userId: null == userId
          ? _value.userId
          : userId // ignore: cast_nullable_to_non_nullable
              as String,
      username: null == username
          ? _value.username
          : username // ignore: cast_nullable_to_non_nullable
              as String,
      avatarEmoji: freezed == avatarEmoji
          ? _value.avatarEmoji
          : avatarEmoji // ignore: cast_nullable_to_non_nullable
              as String?,
      rank: null == rank
          ? _value.rank
          : rank // ignore: cast_nullable_to_non_nullable
              as int,
      ranked: null == ranked
          ? _value.ranked
          : ranked // ignore: cast_nullable_to_non_nullable
              as bool,
      seasonsPlayed: null == seasonsPlayed
          ? _value.seasonsPlayed
          : seasonsPlayed // ignore: cast_nullable_to_non_nullable
              as int,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$StandingEntryImplCopyWith<$Res>
    implements $StandingEntryCopyWith<$Res> {
  factory _$$StandingEntryImplCopyWith(
          _$StandingEntryImpl value, $Res Function(_$StandingEntryImpl) then) =
      __$$StandingEntryImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {@JsonKey(name: 'user_id') String userId,
      String username,
      @JsonKey(name: 'avatar_emoji') String? avatarEmoji,
      int rank,
      bool ranked,
      @JsonKey(name: 'seasons_played') int seasonsPlayed});
}

/// @nodoc
class __$$StandingEntryImplCopyWithImpl<$Res>
    extends _$StandingEntryCopyWithImpl<$Res, _$StandingEntryImpl>
    implements _$$StandingEntryImplCopyWith<$Res> {
  __$$StandingEntryImplCopyWithImpl(
      _$StandingEntryImpl _value, $Res Function(_$StandingEntryImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? userId = null,
    Object? username = null,
    Object? avatarEmoji = freezed,
    Object? rank = null,
    Object? ranked = null,
    Object? seasonsPlayed = null,
  }) {
    return _then(_$StandingEntryImpl(
      userId: null == userId
          ? _value.userId
          : userId // ignore: cast_nullable_to_non_nullable
              as String,
      username: null == username
          ? _value.username
          : username // ignore: cast_nullable_to_non_nullable
              as String,
      avatarEmoji: freezed == avatarEmoji
          ? _value.avatarEmoji
          : avatarEmoji // ignore: cast_nullable_to_non_nullable
              as String?,
      rank: null == rank
          ? _value.rank
          : rank // ignore: cast_nullable_to_non_nullable
              as int,
      ranked: null == ranked
          ? _value.ranked
          : ranked // ignore: cast_nullable_to_non_nullable
              as bool,
      seasonsPlayed: null == seasonsPlayed
          ? _value.seasonsPlayed
          : seasonsPlayed // ignore: cast_nullable_to_non_nullable
              as int,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$StandingEntryImpl implements _StandingEntry {
  const _$StandingEntryImpl(
      {@JsonKey(name: 'user_id') required this.userId,
      required this.username,
      @JsonKey(name: 'avatar_emoji') this.avatarEmoji,
      this.rank = 0,
      this.ranked = false,
      @JsonKey(name: 'seasons_played') this.seasonsPlayed = 0});

  factory _$StandingEntryImpl.fromJson(Map<String, dynamic> json) =>
      _$$StandingEntryImplFromJson(json);

  @override
  @JsonKey(name: 'user_id')
  final String userId;
  @override
  final String username;
  @override
  @JsonKey(name: 'avatar_emoji')
  final String? avatarEmoji;

  /// 1-based, or 0 for a member who has not been through a reveal yet.
  @override
  @JsonKey()
  final int rank;

  /// Whether this member has a place at all. "Has not played here yet" and "placed last" are
  /// different facts, and showing the first as the second says something untrue about a newcomer.
  @override
  @JsonKey()
  final bool ranked;

  /// How much history the place is based on: a place over two seasons and one over ten are not
  /// comparable, so the basis is shown next to the position.
  @override
  @JsonKey(name: 'seasons_played')
  final int seasonsPlayed;

  @override
  String toString() {
    return 'StandingEntry(userId: $userId, username: $username, avatarEmoji: $avatarEmoji, rank: $rank, ranked: $ranked, seasonsPlayed: $seasonsPlayed)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$StandingEntryImpl &&
            (identical(other.userId, userId) || other.userId == userId) &&
            (identical(other.username, username) ||
                other.username == username) &&
            (identical(other.avatarEmoji, avatarEmoji) ||
                other.avatarEmoji == avatarEmoji) &&
            (identical(other.rank, rank) || other.rank == rank) &&
            (identical(other.ranked, ranked) || other.ranked == ranked) &&
            (identical(other.seasonsPlayed, seasonsPlayed) ||
                other.seasonsPlayed == seasonsPlayed));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(
      runtimeType, userId, username, avatarEmoji, rank, ranked, seasonsPlayed);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$StandingEntryImplCopyWith<_$StandingEntryImpl> get copyWith =>
      __$$StandingEntryImplCopyWithImpl<_$StandingEntryImpl>(this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$StandingEntryImplToJson(
      this,
    );
  }
}

abstract class _StandingEntry implements StandingEntry {
  const factory _StandingEntry(
          {@JsonKey(name: 'user_id') required final String userId,
          required final String username,
          @JsonKey(name: 'avatar_emoji') final String? avatarEmoji,
          final int rank,
          final bool ranked,
          @JsonKey(name: 'seasons_played') final int seasonsPlayed}) =
      _$StandingEntryImpl;

  factory _StandingEntry.fromJson(Map<String, dynamic> json) =
      _$StandingEntryImpl.fromJson;

  @override
  @JsonKey(name: 'user_id')
  String get userId;
  @override
  String get username;
  @override
  @JsonKey(name: 'avatar_emoji')
  String? get avatarEmoji;
  @override

  /// 1-based, or 0 for a member who has not been through a reveal yet.
  int get rank;
  @override

  /// Whether this member has a place at all. "Has not played here yet" and "placed last" are
  /// different facts, and showing the first as the second says something untrue about a newcomer.
  bool get ranked;
  @override

  /// How much history the place is based on: a place over two seasons and one over ten are not
  /// comparable, so the basis is shown next to the position.
  @JsonKey(name: 'seasons_played')
  int get seasonsPlayed;
  @override
  @JsonKey(ignore: true)
  _$$StandingEntryImplCopyWith<_$StandingEntryImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

GuessStanding _$GuessStandingFromJson(Map<String, dynamic> json) {
  return _GuessStanding.fromJson(json);
}

/// @nodoc
mixin _$GuessStanding {
  bool get available => throw _privateConstructorUsedError;
  String get reason => throw _privateConstructorUsedError;
  List<StandingEntry> get members => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $GuessStandingCopyWith<GuessStanding> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $GuessStandingCopyWith<$Res> {
  factory $GuessStandingCopyWith(
          GuessStanding value, $Res Function(GuessStanding) then) =
      _$GuessStandingCopyWithImpl<$Res, GuessStanding>;
  @useResult
  $Res call({bool available, String reason, List<StandingEntry> members});
}

/// @nodoc
class _$GuessStandingCopyWithImpl<$Res, $Val extends GuessStanding>
    implements $GuessStandingCopyWith<$Res> {
  _$GuessStandingCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? available = null,
    Object? reason = null,
    Object? members = null,
  }) {
    return _then(_value.copyWith(
      available: null == available
          ? _value.available
          : available // ignore: cast_nullable_to_non_nullable
              as bool,
      reason: null == reason
          ? _value.reason
          : reason // ignore: cast_nullable_to_non_nullable
              as String,
      members: null == members
          ? _value.members
          : members // ignore: cast_nullable_to_non_nullable
              as List<StandingEntry>,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$GuessStandingImplCopyWith<$Res>
    implements $GuessStandingCopyWith<$Res> {
  factory _$$GuessStandingImplCopyWith(
          _$GuessStandingImpl value, $Res Function(_$GuessStandingImpl) then) =
      __$$GuessStandingImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call({bool available, String reason, List<StandingEntry> members});
}

/// @nodoc
class __$$GuessStandingImplCopyWithImpl<$Res>
    extends _$GuessStandingCopyWithImpl<$Res, _$GuessStandingImpl>
    implements _$$GuessStandingImplCopyWith<$Res> {
  __$$GuessStandingImplCopyWithImpl(
      _$GuessStandingImpl _value, $Res Function(_$GuessStandingImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? available = null,
    Object? reason = null,
    Object? members = null,
  }) {
    return _then(_$GuessStandingImpl(
      available: null == available
          ? _value.available
          : available // ignore: cast_nullable_to_non_nullable
              as bool,
      reason: null == reason
          ? _value.reason
          : reason // ignore: cast_nullable_to_non_nullable
              as String,
      members: null == members
          ? _value._members
          : members // ignore: cast_nullable_to_non_nullable
              as List<StandingEntry>,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$GuessStandingImpl implements _GuessStanding {
  const _$GuessStandingImpl(
      {this.available = false,
      this.reason = '',
      final List<StandingEntry> members = const []})
      : _members = members;

  factory _$GuessStandingImpl.fromJson(Map<String, dynamic> json) =>
      _$$GuessStandingImplFromJson(json);

  @override
  @JsonKey()
  final bool available;
  @override
  @JsonKey()
  final String reason;
  final List<StandingEntry> _members;
  @override
  @JsonKey()
  List<StandingEntry> get members {
    if (_members is EqualUnmodifiableListView) return _members;
    // ignore: implicit_dynamic_type
    return EqualUnmodifiableListView(_members);
  }

  @override
  String toString() {
    return 'GuessStanding(available: $available, reason: $reason, members: $members)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$GuessStandingImpl &&
            (identical(other.available, available) ||
                other.available == available) &&
            (identical(other.reason, reason) || other.reason == reason) &&
            const DeepCollectionEquality().equals(other._members, _members));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(runtimeType, available, reason,
      const DeepCollectionEquality().hash(_members));

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$GuessStandingImplCopyWith<_$GuessStandingImpl> get copyWith =>
      __$$GuessStandingImplCopyWithImpl<_$GuessStandingImpl>(this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$GuessStandingImplToJson(
      this,
    );
  }
}

abstract class _GuessStanding implements GuessStanding {
  const factory _GuessStanding(
      {final bool available,
      final String reason,
      final List<StandingEntry> members}) = _$GuessStandingImpl;

  factory _GuessStanding.fromJson(Map<String, dynamic> json) =
      _$GuessStandingImpl.fromJson;

  @override
  bool get available;
  @override
  String get reason;
  @override
  List<StandingEntry> get members;
  @override
  @JsonKey(ignore: true)
  _$$GuessStandingImplCopyWith<_$GuessStandingImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

Chronicle _$ChronicleFromJson(Map<String, dynamic> json) {
  return _Chronicle.fromJson(json);
}

/// @nodoc
mixin _$Chronicle {
  List<ChronicleSeason> get seasons => throw _privateConstructorUsedError;
  GuessStanding get standing => throw _privateConstructorUsedError;
  @JsonKey(name: 'seasons_total')
  int get seasonsTotal => throw _privateConstructorUsedError;
  @JsonKey(name: 'seasons_shown')
  int get seasonsShown => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $ChronicleCopyWith<Chronicle> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $ChronicleCopyWith<$Res> {
  factory $ChronicleCopyWith(Chronicle value, $Res Function(Chronicle) then) =
      _$ChronicleCopyWithImpl<$Res, Chronicle>;
  @useResult
  $Res call(
      {List<ChronicleSeason> seasons,
      GuessStanding standing,
      @JsonKey(name: 'seasons_total') int seasonsTotal,
      @JsonKey(name: 'seasons_shown') int seasonsShown});

  $GuessStandingCopyWith<$Res> get standing;
}

/// @nodoc
class _$ChronicleCopyWithImpl<$Res, $Val extends Chronicle>
    implements $ChronicleCopyWith<$Res> {
  _$ChronicleCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? seasons = null,
    Object? standing = null,
    Object? seasonsTotal = null,
    Object? seasonsShown = null,
  }) {
    return _then(_value.copyWith(
      seasons: null == seasons
          ? _value.seasons
          : seasons // ignore: cast_nullable_to_non_nullable
              as List<ChronicleSeason>,
      standing: null == standing
          ? _value.standing
          : standing // ignore: cast_nullable_to_non_nullable
              as GuessStanding,
      seasonsTotal: null == seasonsTotal
          ? _value.seasonsTotal
          : seasonsTotal // ignore: cast_nullable_to_non_nullable
              as int,
      seasonsShown: null == seasonsShown
          ? _value.seasonsShown
          : seasonsShown // ignore: cast_nullable_to_non_nullable
              as int,
    ) as $Val);
  }

  @override
  @pragma('vm:prefer-inline')
  $GuessStandingCopyWith<$Res> get standing {
    return $GuessStandingCopyWith<$Res>(_value.standing, (value) {
      return _then(_value.copyWith(standing: value) as $Val);
    });
  }
}

/// @nodoc
abstract class _$$ChronicleImplCopyWith<$Res>
    implements $ChronicleCopyWith<$Res> {
  factory _$$ChronicleImplCopyWith(
          _$ChronicleImpl value, $Res Function(_$ChronicleImpl) then) =
      __$$ChronicleImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {List<ChronicleSeason> seasons,
      GuessStanding standing,
      @JsonKey(name: 'seasons_total') int seasonsTotal,
      @JsonKey(name: 'seasons_shown') int seasonsShown});

  @override
  $GuessStandingCopyWith<$Res> get standing;
}

/// @nodoc
class __$$ChronicleImplCopyWithImpl<$Res>
    extends _$ChronicleCopyWithImpl<$Res, _$ChronicleImpl>
    implements _$$ChronicleImplCopyWith<$Res> {
  __$$ChronicleImplCopyWithImpl(
      _$ChronicleImpl _value, $Res Function(_$ChronicleImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? seasons = null,
    Object? standing = null,
    Object? seasonsTotal = null,
    Object? seasonsShown = null,
  }) {
    return _then(_$ChronicleImpl(
      seasons: null == seasons
          ? _value._seasons
          : seasons // ignore: cast_nullable_to_non_nullable
              as List<ChronicleSeason>,
      standing: null == standing
          ? _value.standing
          : standing // ignore: cast_nullable_to_non_nullable
              as GuessStanding,
      seasonsTotal: null == seasonsTotal
          ? _value.seasonsTotal
          : seasonsTotal // ignore: cast_nullable_to_non_nullable
              as int,
      seasonsShown: null == seasonsShown
          ? _value.seasonsShown
          : seasonsShown // ignore: cast_nullable_to_non_nullable
              as int,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$ChronicleImpl implements _Chronicle {
  const _$ChronicleImpl(
      {final List<ChronicleSeason> seasons = const [],
      this.standing = const GuessStanding(),
      @JsonKey(name: 'seasons_total') this.seasonsTotal = 0,
      @JsonKey(name: 'seasons_shown') this.seasonsShown = 0})
      : _seasons = seasons;

  factory _$ChronicleImpl.fromJson(Map<String, dynamic> json) =>
      _$$ChronicleImplFromJson(json);

  final List<ChronicleSeason> _seasons;
  @override
  @JsonKey()
  List<ChronicleSeason> get seasons {
    if (_seasons is EqualUnmodifiableListView) return _seasons;
    // ignore: implicit_dynamic_type
    return EqualUnmodifiableListView(_seasons);
  }

  @override
  @JsonKey()
  final GuessStanding standing;
  @override
  @JsonKey(name: 'seasons_total')
  final int seasonsTotal;
  @override
  @JsonKey(name: 'seasons_shown')
  final int seasonsShown;

  @override
  String toString() {
    return 'Chronicle(seasons: $seasons, standing: $standing, seasonsTotal: $seasonsTotal, seasonsShown: $seasonsShown)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$ChronicleImpl &&
            const DeepCollectionEquality().equals(other._seasons, _seasons) &&
            (identical(other.standing, standing) ||
                other.standing == standing) &&
            (identical(other.seasonsTotal, seasonsTotal) ||
                other.seasonsTotal == seasonsTotal) &&
            (identical(other.seasonsShown, seasonsShown) ||
                other.seasonsShown == seasonsShown));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(
      runtimeType,
      const DeepCollectionEquality().hash(_seasons),
      standing,
      seasonsTotal,
      seasonsShown);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$ChronicleImplCopyWith<_$ChronicleImpl> get copyWith =>
      __$$ChronicleImplCopyWithImpl<_$ChronicleImpl>(this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$ChronicleImplToJson(
      this,
    );
  }
}

abstract class _Chronicle implements Chronicle {
  const factory _Chronicle(
          {final List<ChronicleSeason> seasons,
          final GuessStanding standing,
          @JsonKey(name: 'seasons_total') final int seasonsTotal,
          @JsonKey(name: 'seasons_shown') final int seasonsShown}) =
      _$ChronicleImpl;

  factory _Chronicle.fromJson(Map<String, dynamic> json) =
      _$ChronicleImpl.fromJson;

  @override
  List<ChronicleSeason> get seasons;
  @override
  GuessStanding get standing;
  @override
  @JsonKey(name: 'seasons_total')
  int get seasonsTotal;
  @override
  @JsonKey(name: 'seasons_shown')
  int get seasonsShown;
  @override
  @JsonKey(ignore: true)
  _$$ChronicleImplCopyWith<_$ChronicleImpl> get copyWith =>
      throw _privateConstructorUsedError;
}
